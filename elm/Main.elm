port module Main exposing (main)

import Api
import Browser
import Browser.Events
import Browser.Navigation as Nav
import Debug
import Element exposing (Attribute, Element)
import Element.Events as Events
import Element.Input as Input
import Html exposing (Html)
import Html.Attributes as Attr
import Http
import Page.ConfirmForgotPassword as ConfirmForgotPassword
import Page.ConfirmSignUp as ConfirmSignUp
import Page.ForgotPassword as ForgotPassword
import Page.Kifu as Kifu
import Page.MyPage as MyPage
import Page.ResendConfirm as ResendConfirm
import Page.SignIn as SignIn
import Page.SignUp as SignUp
import Page.Upload as Upload
import Proto.Api as PB
import Route exposing (Route)
import Style
import Url exposing (Url)


port storeToken : String -> Cmd msg


port storeTokens : ( String, String ) -> Cmd msg


port removeTokens : () -> Cmd msg


port updateBoard : ( String, String ) -> Cmd msg


type alias Flags =
    { token : Maybe String
    , refreshToken : Maybe String
    , windowWidth : Int
    , windowHeight : Int
    }


type Msg
    = UrlRequest Browser.UrlRequest
    | UrlChanged Url
    | OnResize Int Int
    | SignUpMsg SignUp.Msg
    | ConfirmSignUpMsg ConfirmSignUp.Msg
    | ResendConfirmMsg ResendConfirm.Msg
    | SignInMsg SignIn.Msg
    | ForgotPasswordMsg ForgotPassword.Msg
    | ConfirmForgotPasswordMsg ConfirmForgotPassword.Msg
    | ApiResponse Api.Request Api.Response
    | UploadMsg Upload.Msg
    | KifuMsg Kifu.Msg
    | HelloRequest
    | Logout
    | NOP


type PreviousState
    = PrevNone
    | PrevRequest Api.Request


type alias Model =
    { key : Nav.Key
    , route : Route
    , windowSize : ( Int, Int )
    , signUpModel : SignUp.Model
    , confirmSignUpModel : ConfirmSignUp.Model
    , resendConfirmModel : ResendConfirm.Model
    , signInModel : SignIn.Model
    , forgotPasswordModel : ForgotPassword.Model
    , confirmForgotPasswordModel : ConfirmForgotPassword.Model
    , authToken : Maybe PB.SignInResponse
    , errorMessage : Maybe String
    , prevState : PreviousState
    , recentKifu : List PB.RecentKifuResponse_Kifu
    , kifuModel : Kifu.Model
    , uploadModel : Upload.Model
    }


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = UrlRequest
        }


init : Flags -> Url -> Nav.Key -> ( Model, Cmd Msg )
init flags url key =
    let
        authToken =
            case ( flags.token, flags.refreshToken ) of
                ( Just token, Just refreshToken ) ->
                    Just { token = token, refreshToken = refreshToken }

                _ ->
                    Nothing
    in
    ( { key = key
      , route = Route.fromUrl url
      , windowSize = ( flags.windowWidth, flags.windowHeight )
      , signUpModel = SignUp.init
      , confirmSignUpModel = ConfirmSignUp.init
      , resendConfirmModel = ResendConfirm.init
      , signInModel = SignIn.init
      , forgotPasswordModel = ForgotPassword.init
      , confirmForgotPasswordModel = ConfirmForgotPassword.init
      , authToken = authToken
      , errorMessage = Nothing
      , prevState = PrevNone
      , recentKifu = []
      , kifuModel = Kifu.init
      , uploadModel = Upload.init False
      }
    , Nav.pushUrl key (Url.toString url)
    )


httpErrorToString : Http.Error -> String
httpErrorToString err =
    case err of
        Http.BadUrl str ->
            "BadUrl: " ++ str

        Http.Timeout ->
            "Timeout"

        Http.NetworkError ->
            "NetworkError"

        Http.BadStatus i ->
            "BadStatus: " ++ String.fromInt i

        Http.BadBody str ->
            "BadBody: " ++ str


signInAndReturn : Model -> Maybe PB.SignInResponse -> String -> Maybe String -> ( Model, Cmd Msg )
signInAndReturn model authToken token refreshToken =
    let
        model_ =
            { model
                | authToken = authToken
                , prevState = PrevNone
            }

        cmdStoreTokens =
            case refreshToken of
                Just refreshToken_ ->
                    storeTokens ( token, refreshToken_ )

                Nothing ->
                    storeToken token
    in
    case model.prevState of
        PrevRequest req ->
            ( model_
            , Cmd.batch
                [ cmdStoreTokens
                , Api.request ApiResponse (Just token) req
                ]
            )

        PrevNone ->
            ( model_
            , Cmd.batch
                [ cmdStoreTokens
                , Nav.pushUrl model.key (Route.path Route.MyPage)
                ]
            )


authorizedResponse :
    Model
    -> Api.Request
    -> Result Api.Error a
    -> (a -> ( Model, Cmd Msg ))
    -> ( Model, Cmd Msg )
authorizedResponse model req result f =
    case result of
        Ok r ->
            f r

        Err Api.ErrorUnauthorized ->
            case model.authToken of
                Just t ->
                    ( { model | prevState = PrevRequest req }
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestTokenRefresh { refreshToken = t.refreshToken }
                    )

                Nothing ->
                    ( { model | prevState = PrevRequest req }
                    , Nav.pushUrl model.key (Route.path Route.SignIn)
                    )

        Err err ->
            Debug.log (Debug.toString err) ( model, Cmd.none )


elem : List a -> Int -> Maybe a
elem list n =
    List.head <| List.drop n list


updateKifuPage : Model -> Kifu.Model -> String -> Int -> ( Model, Cmd Msg )
updateKifuPage model kifuModel kifuId seq =
    if seq >= 0 then
        case elem kifuModel.kifu.steps seq of
            Just step ->
                ( { model | kifuModel = { kifuModel | curSeq = seq } }
                , updateBoard ( "shogi", step.position )
                )

            Nothing ->
                ( model
                , Nav.pushUrl
                    model.key
                    (Route.path <| Route.Kifu kifuId kifuModel.curSeq)
                )

    else
        ( model
        , Nav.pushUrl
            model.key
            (Route.path <| Route.Kifu kifuId 0)
        )


apiResponse : Model -> Api.Request -> Api.Response -> ( Model, Cmd Msg )
apiResponse model req res =
    case res of
        Api.AuthResponse result ->
            case result of
                Ok v ->
                    case v.authResponseSelect of
                        PB.ResponseSignUp r ->
                            let
                                m =
                                    model.confirmSignUpModel
                            in
                            ( { model
                                | confirmSignUpModel = { m | signUpResponse = Just r }
                              }
                            , Nav.pushUrl model.key (Route.path Route.ConfirmSignUp)
                            )

                        PB.ResponseConfirmSignUp _ ->
                            let
                                m =
                                    model.confirmSignUpModel
                            in
                            ( { model | confirmSignUpModel = ConfirmSignUp.init }
                            , Nav.pushUrl model.key (Route.path Route.Index)
                            )

                        PB.ResponseForgotPassword r ->
                            let
                                m =
                                    model.confirmForgotPasswordModel
                            in
                            ( { model
                                | confirmForgotPasswordModel = { m | forgotPasswordResponse = Just r }
                              }
                            , Nav.pushUrl model.key (Route.path Route.ConfirmForgotPassword)
                            )

                        PB.ResponseConfirmForgotPassword _ ->
                            ( model
                            , Nav.pushUrl model.key (Route.path Route.Index)
                            )

                        PB.ResponseSignIn r ->
                            signInAndReturn model (Just r) r.token (Just r.refreshToken)

                        PB.ResponseTokenRefresh r ->
                            signInAndReturn model
                                (Maybe.map
                                    (\t -> { t | token = r.token })
                                    model.authToken
                                )
                                r.token
                                Nothing

                        PB.AuthResponseSelectUnspecified ->
                            ( model, Cmd.none )

                Err Api.ErrorUnauthorized ->
                    ( { model
                        | errorMessage = Just "logout"
                        , authToken = Nothing
                      }
                    , Cmd.none
                    )

                Err err ->
                    ( { model
                        | errorMessage = Just <| Api.errorToString err
                      }
                    , Cmd.none
                    )

        Api.KifuResponse result ->
            authorizedResponse model req result <|
                \kifuRes ->
                    case kifuRes.kifuResponseSelect of
                        PB.ResponseRecentKifu r ->
                            ( { model | recentKifu = r.kifus }
                            , Cmd.none
                            )

                        PB.ResponsePostKifu r ->
                            ( model
                            , Nav.pushUrl model.key <|
                                Route.path <|
                                    if model.uploadModel.repeat then
                                        Route.Upload

                                    else
                                        Route.Index
                            )

                        PB.ResponseGetKifu r ->
                            let
                                curSeq =
                                    case model.route of
                                        Route.Kifu _ seq ->
                                            seq

                                        _ ->
                                            0

                                kifu =
                                    { r | steps = List.sortBy (\s -> s.seq) r.steps }

                                model_ =
                                    { model
                                        | kifuModel =
                                            { kifu = kifu
                                            , curSeq = curSeq
                                            , len = List.length kifu.steps
                                            }
                                    }
                            in
                            updateKifuPage model_ model_.kifuModel r.kifuId curSeq

                        _ ->
                            ( model, Cmd.none )

        Api.HelloResponse result ->
            authorizedResponse model req result <|
                \r ->
                    let
                        _ =
                            Debug.log "HelloResponse" r
                    in
                    ( model, Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    let
        authToken =
            Maybe.map (\at -> at.token) model.authToken
    in
    case msg of
        UrlRequest urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            let
                model_ =
                    { model
                        | route = Route.fromUrl url
                        , errorMessage = Nothing
                    }
            in
            case model_.route of
                Route.SignUp ->
                    ( { model_ | signUpModel = SignUp.init }
                    , Cmd.none
                    )

                Route.SignIn ->
                    ( { model_ | signInModel = SignIn.init }
                    , Cmd.none
                    )

                Route.ConfirmSignUp ->
                    ( { model_ | confirmSignUpModel = ConfirmSignUp.init }
                    , Cmd.none
                    )

                Route.ForgotPassword ->
                    ( { model_ | forgotPasswordModel = ForgotPassword.init }
                    , Cmd.none
                    )

                Route.MyPage ->
                    ( model_
                    , Api.request ApiResponse authToken <|
                        Api.KifuRequest <|
                            PB.KifuRequest <|
                                PB.RequestRecentKifu
                                    { limit = 10
                                    }
                    )

                Route.Kifu kifuId seq ->
                    let
                        km =
                            model_.kifuModel
                    in
                    if km.kifu.kifuId == kifuId then
                        updateKifuPage model_ model_.kifuModel kifuId seq

                    else
                        ( model_
                        , Api.request ApiResponse authToken <|
                            Api.KifuRequest <|
                                PB.KifuRequest <|
                                    PB.RequestGetKifu
                                        { kifuId = kifuId
                                        }
                        )

                Route.Upload ->
                    ( { model_ | uploadModel = Upload.init model.uploadModel.repeat }
                    , Cmd.none
                    )

                _ ->
                    ( model_, Cmd.none )

        SignUpMsg m ->
            case m of
                SignUp.Submit ->
                    ( model
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestSignUp model.signUpModel.params
                    )

                _ ->
                    ( { model | signUpModel = SignUp.update m model.signUpModel }
                    , Cmd.none
                    )

        ConfirmSignUpMsg m ->
            case m of
                ConfirmSignUp.Submit ->
                    ( model
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestConfirmSignUp model.confirmSignUpModel.params
                    )

                _ ->
                    ( { model | confirmSignUpModel = ConfirmSignUp.update m model.confirmSignUpModel }
                    , Cmd.none
                    )

        SignInMsg m ->
            case m of
                SignIn.Submit ->
                    ( model
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestSignIn model.signInModel.params
                    )

                _ ->
                    ( { model | signInModel = SignIn.update m model.signInModel }
                    , Cmd.none
                    )

        ForgotPasswordMsg m ->
            case m of
                ForgotPassword.Submit ->
                    ( model
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestForgotPassword model.forgotPasswordModel
                    )

                _ ->
                    ( { model | forgotPasswordModel = ForgotPassword.update m model.forgotPasswordModel }
                    , Cmd.none
                    )

        ConfirmForgotPasswordMsg m ->
            case m of
                ConfirmForgotPassword.Submit ->
                    ( model
                    , Api.request ApiResponse Nothing <|
                        Api.AuthRequest <|
                            PB.AuthRequest <|
                                PB.RequestConfirmForgotPassword model.confirmForgotPasswordModel.params
                    )

                _ ->
                    ( { model | confirmForgotPasswordModel = ConfirmForgotPassword.update m model.confirmForgotPasswordModel }
                    , Cmd.none
                    )

        ApiResponse req res ->
            apiResponse model req res

        UploadMsg uploadMsg ->
            case uploadMsg of
                Upload.Submit ->
                    ( model
                    , Api.request ApiResponse authToken <|
                        Api.KifuRequest <|
                            PB.KifuRequest <|
                                PB.RequestPostKifu model.uploadModel.request
                    )

                _ ->
                    ( { model | uploadModel = Upload.update uploadMsg model.uploadModel }
                    , Cmd.none
                    )

        KifuMsg kifuMsg ->
            case kifuMsg of
                Kifu.UpdateBoard kifuId seq ->
                    ( model
                    , Nav.pushUrl model.key (Route.path <| Route.Kifu kifuId seq)
                    )

        HelloRequest ->
            ( model
            , Api.request ApiResponse authToken Api.HelloRequest
            )

        OnResize w h ->
            ( { model | windowSize = ( w, h ) }, Cmd.none )

        Logout ->
            ( model, removeTokens () )

        _ ->
            let
                _ =
                    Debug.log (Debug.toString msg) msg
            in
            ( model, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Browser.Events.onResize OnResize


routeToTitle : Route -> String
routeToTitle route =
    case route of
        Route.Index ->
            "Index"

        Route.SignUp ->
            "SignUp"

        Route.ConfirmSignUp ->
            "Confirm SignUp"

        Route.ResendConfirm ->
            "Resend confirm code"

        Route.SignIn ->
            "SignIn"

        Route.ForgotPassword ->
            "Forgot password"

        Route.ConfirmForgotPassword ->
            "Confirm forgot password"

        Route.MyPage ->
            "MyPage"

        Route.Upload ->
            "Upload"

        Route.Kifu kifuId seq ->
            String.concat
                [ "棋譜: "
                , String.fromInt seq
                , "手目 ("
                , kifuId
                , ")"
                ]

        Route.NotFound ->
            "NotFound"


content : Model -> Element Msg
content model =
    case model.route of
        Route.SignUp ->
            SignUp.view SignUpMsg model.signUpModel

        Route.ConfirmSignUp ->
            ConfirmSignUp.view ConfirmSignUpMsg model.confirmSignUpModel

        Route.ResendConfirm ->
            ResendConfirm.view ResendConfirmMsg model.resendConfirmModel

        Route.SignIn ->
            SignIn.view SignInMsg model.signInModel

        Route.ForgotPassword ->
            ForgotPassword.view ForgotPasswordMsg model.forgotPasswordModel

        Route.ConfirmForgotPassword ->
            ConfirmForgotPassword.view ConfirmForgotPasswordMsg model.confirmForgotPasswordModel

        Route.MyPage ->
            MyPage.view model.recentKifu

        Route.Upload ->
            Upload.view UploadMsg model.uploadModel

        Route.Kifu kifuId seq ->
            Kifu.view KifuMsg model.kifuModel

        Route.NotFound ->
            Element.column []
                [ Element.text "NotFound"
                , Element.html <|
                    Html.ul []
                        [ Html.a [ Attr.href "./" ] [ Html.text "Index" ]
                        ]
                ]

        Route.Index ->
            let
                url =
                    "http://shineleckoma.web.fc2.com/"
            in
            Element.column Style.mainColumn
                [ Element.text "駒画像はしんえれ外部駒のものを使用しています。"
                , Element.link [] { url = url, label = Element.text url }
                ]


headerAttrs : List (Attribute msg)
headerAttrs =
    [ Element.spacing 10
    ]


userInfo : Model -> Element Msg
userInfo model =
    if model.authToken == Nothing then
        Element.row headerAttrs
            [ Element.link [] { url = Route.path Route.SignUp, label = Element.text "Sign up" }
            , Style.border
            , Element.link [] { url = Route.path Route.SignIn, label = Element.text "Sign in" }
            ]

    else
        Element.row headerAttrs
            [ Element.link [] { url = Route.path Route.MyPage, label = Element.text "My page" }
            , Style.border
            , Input.button [] { onPress = Just Logout, label = Element.text "Logout" }
            ]


header : Model -> Element Msg
header model =
    Element.row headerAttrs
        [ Element.link [] { url = Route.path Route.Index, label = Element.text "Index" }
        , Style.border
        , userInfo model
        ]


view : Model -> Browser.Document Msg
view model =
    { title = routeToTitle model.route
    , body =
        [ Element.layout [ Element.padding 10 ] <|
            Element.column [ Element.spacing 20 ]
                [ header model
                , Element.el [] <| Element.text <| Maybe.withDefault "" model.errorMessage
                , content model
                ]
        ]
    }
