port module Main exposing (main)

import Api
import Browser
import Browser.Navigation as Nav
import Debug
import Element exposing (Element)
import Element.Events as Events
import Html exposing (Html)
import Html.Attributes as Attr
import Http
import Page.ConfirmForgotPassword as ConfirmForgotPassword
import Page.ConfirmSignUp as ConfirmSignUp
import Page.ForgotPassword as ForgotPassword
import Page.ResendConfirm as ResendConfirm
import Page.SignIn as SignIn
import Page.SignUp as SignUp
import Proto.Api as PB
import Url exposing (Url)
import Url.Builder as UrlBuilder
import Url.Parser as UrlParser exposing ((</>), Parser, s)


port storeToken : String -> Cmd msg


port storeTokens : ( String, String ) -> Cmd msg


type alias Flags =
    { token : Maybe String
    , refreshToken : Maybe String
    }


type Msg
    = LinkClicked Browser.UrlRequest
    | UrlChanged Url
    | SignUpMsg SignUp.Msg
    | ConfirmSignUpMsg ConfirmSignUp.Msg
    | ResendConfirmMsg ResendConfirm.Msg
    | SignInMsg SignIn.Msg
    | ForgotPasswordMsg ForgotPassword.Msg
    | ConfirmForgotPasswordMsg ConfirmForgotPassword.Msg
    | ApiResponse Api.Request Api.Response
    | HelloRequest
    | NOP


type Route
    = Index
    | SignUp
    | ConfirmSignUp
    | ResendConfirm
    | ForgotPassword
    | ConfirmForgotPassword
    | SignIn
    | MyPage
    | NotFound


routeParser : Parser (Route -> a) a
routeParser =
    UrlParser.oneOf
        [ UrlParser.map Index UrlParser.top
        , UrlParser.map Index <| s "index.html"
        , UrlParser.map SignUp <| s "signup"
        , UrlParser.map ConfirmSignUp <| s "confirm_signup"
        , UrlParser.map ResendConfirm <| s "resend_confirm"
        , UrlParser.map SignIn <| s "signin"
        , UrlParser.map ForgotPassword <| s "forgot_password"
        , UrlParser.map ConfirmForgotPassword <| s "confirm_forgot_password"
        , UrlParser.map MyPage <| s "my"
        ]


toRoute : Url -> Route
toRoute url =
    Maybe.withDefault NotFound (UrlParser.parse routeParser url)


routeToPath : Route -> String
routeToPath route =
    case route of
        Index ->
            UrlBuilder.absolute [] []

        SignUp ->
            UrlBuilder.absolute [ "signup" ] []

        ConfirmSignUp ->
            UrlBuilder.absolute [ "confirm_signup" ] []

        ResendConfirm ->
            UrlBuilder.absolute [ "resend_confirm" ] []

        SignIn ->
            UrlBuilder.absolute [ "signin" ] []

        ForgotPassword ->
            UrlBuilder.absolute [ "forgot_password" ] []

        ConfirmForgotPassword ->
            UrlBuilder.absolute [ "confirm_forgot_password" ] []

        MyPage ->
            UrlBuilder.absolute [ "my" ] []

        NotFound ->
            UrlBuilder.absolute [] []


type PreviousState
    = PrevNone
    | PrevRequest Api.Request


type alias Model =
    { key : Nav.Key
    , route : Route
    , signUpModel : SignUp.Model
    , confirmSignUpModel : ConfirmSignUp.Model
    , resendConfirmModel : ResendConfirm.Model
    , signInModel : SignIn.Model
    , forgotPasswordModel : ForgotPassword.Model
    , confirmForgotPasswordModel : ConfirmForgotPassword.Model
    , authToken : Maybe PB.SignInResponse
    , errorMessage : Maybe String
    , prevState : PreviousState

    --, recentKifu:PB.RecentKifuResponse
    }


main : Program Flags Model Msg
main =
    Browser.application
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
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
      , route = toRoute url
      , signUpModel = SignUp.init
      , confirmSignUpModel = ConfirmSignUp.init
      , resendConfirmModel = ResendConfirm.init
      , signInModel = SignIn.init
      , forgotPasswordModel = ForgotPassword.init
      , confirmForgotPasswordModel = ConfirmForgotPassword.init
      , authToken = authToken
      , errorMessage = Nothing
      , prevState = PrevNone
      }
    , Cmd.none
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
                , Nav.pushUrl model.key (routeToPath Index)
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
                    , Nav.pushUrl model.key (routeToPath SignIn)
                    )

        Err err ->
            Debug.log (Debug.toString err) ( model, Cmd.none )


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
                            , Nav.pushUrl model.key (routeToPath ConfirmSignUp)
                            )

                        PB.ResponseConfirmSignUp _ ->
                            let
                                m =
                                    model.confirmSignUpModel
                            in
                            ( { model | confirmSignUpModel = ConfirmSignUp.init }
                            , Nav.pushUrl model.key (routeToPath Index)
                            )

                        PB.ResponseForgotPassword r ->
                            let
                                m =
                                    model.confirmForgotPasswordModel
                            in
                            ( { model
                                | confirmForgotPasswordModel = { m | forgotPasswordResponse = Just r }
                              }
                            , Nav.pushUrl model.key (routeToPath ConfirmForgotPassword)
                            )

                        PB.ResponseConfirmForgotPassword _ ->
                            ( model
                            , Nav.pushUrl model.key (routeToPath Index)
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
            -- TODO
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
    case msg of
        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            ( { model
                | route = toRoute url
                , signUpModel = SignUp.init
                , signInModel = SignIn.init
                , errorMessage = Nothing
              }
            , Cmd.none
            )

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

                ConfirmSignUp.ResendCode ->
                    ( model
                    , Nav.pushUrl model.key (routeToPath ResendConfirm)
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

                SignIn.ForgotPassword ->
                    ( model
                    , Nav.pushUrl model.key (routeToPath ForgotPassword)
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

        HelloRequest ->
            ( model
            , Api.request ApiResponse (Maybe.map (\at -> at.token) model.authToken) Api.HelloRequest
            )

        _ ->
            let
                _ =
                    Debug.log (Debug.toString msg) msg
            in
            ( model, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none


routeToTitle : Route -> String
routeToTitle route =
    case route of
        Index ->
            "Index"

        SignUp ->
            "SignUp"

        ConfirmSignUp ->
            "Confirm SignUp"

        ResendConfirm ->
            "Resend confirm code"

        SignIn ->
            "SignIn"

        ForgotPassword ->
            "Forgot password"

        ConfirmForgotPassword ->
            "Confirm forgot password"

        MyPage ->
            "MyPage"

        NotFound ->
            "NotFound"


content : Model -> Element Msg
content model =
    case model.route of
        SignUp ->
            SignUp.view SignUpMsg model.signUpModel

        ConfirmSignUp ->
            ConfirmSignUp.view ConfirmSignUpMsg model.confirmSignUpModel

        ResendConfirm ->
            ResendConfirm.view ResendConfirmMsg model.resendConfirmModel

        SignIn ->
            SignIn.view SignInMsg model.signInModel

        ForgotPassword ->
            ForgotPassword.view ForgotPasswordMsg model.forgotPasswordModel

        ConfirmForgotPassword ->
            ConfirmForgotPassword.view ConfirmForgotPasswordMsg model.confirmForgotPasswordModel

        NotFound ->
            Element.column []
                [ Element.text "NotFound"
                , Element.html <|
                    Html.ul []
                        [ Html.a [ Attr.href "./" ] [ Html.text "Index" ]
                        ]
                ]

        Index ->
            Element.column []
                [ Element.text "index"
                , Element.link
                    [ Events.onClick HelloRequest ]
                    { url = "", label = Element.text "test" }
                ]

        MyPage ->
            Element.column []
                [ Element.text "Recent kifu"
                , Element.link
                    [ Events.onClick HelloRequest ]
                    { url = "", label = Element.text "test" }
                ]


view : Model -> Browser.Document Msg
view model =
    { title = routeToTitle model.route
    , body =
        [ Element.layout [] <|
            Element.column []
                [ Element.row [ Element.spaceEvenly ]
                    [ Element.link [] { url = routeToPath Index, label = Element.text "Header" }
                    , Element.text "|"
                    , Element.link [] { url = routeToPath SignUp, label = Element.text "Sign up" }
                    , Element.text "|"
                    , Element.link [] { url = routeToPath SignIn, label = Element.text "Sign in" }
                    ]
                , Element.el [] <| Element.text <| Maybe.withDefault "" model.errorMessage
                , content model
                , Element.text "footer"
                ]
        ]
    }
