port module Main exposing (main)

--import Debug

import Api
import Browser
import Browser.Events
import Browser.Navigation as Nav
import Element exposing (Attribute, Element)
import Element.Events as Events
import Element.Input as Input
import Html exposing (Html)
import Html.Attributes as Attr
import Http
import Page.Kifu as Kifu
import Page.MyPage as MyPage
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
    , authClientId : String
    , authRedirectURI : String
    }


type Msg
    = UrlRequest Browser.UrlRequest
    | UrlChanged Url
    | OnResize Int Int
    | ApiResponse Api.Request Api.Response
    | UploadMsg Upload.Msg
    | KifuMsg Kifu.Msg
    | Logout
    | NOP


type PreviousState
    = PrevNone
    | PrevRequest Api.Request


type alias Model =
    { key : Nav.Key
    , route : Route
    , windowSize : ( Int, Int )
    , authToken : Maybe PB.SignInResponse
    , errorMessage : Maybe String
    , prevState : PreviousState
    , recentKifu : List PB.RecentKifuResponse_Kifu
    , kifuModel : Kifu.Model
    , uploadModel : Upload.Model
    , loginFormURL : String
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

        loginFormURL =
            String.concat
                [ "https://kansousenauth.wagahai.info/login?response_type=code&client_id="
                , flags.authClientId
                , "&redirect_uri="
                , flags.authRedirectURI
                ]
    in
    ( { key = key
      , route = Route.fromUrl url
      , windowSize = ( flags.windowWidth, flags.windowHeight )
      , authToken = authToken
      , errorMessage = Nothing
      , prevState = PrevNone
      , recentKifu = []
      , kifuModel = Kifu.init
      , uploadModel = Upload.init False
      , loginFormURL = loginFormURL
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
                      -- TODO
                    , Cmd.none
                    )

                Nothing ->
                    ( { model | prevState = PrevRequest req }
                      -- TODO redirect to signin
                    , Cmd.none
                    )

        Err err ->
            --Debug.log (Debug.toString err) ( model, Cmd.none )
            ( model, Cmd.none )


elem : List a -> Int -> Maybe a
elem list n =
    List.head <| List.drop n list


updateKifuPage : Maybe String -> Kifu.Model -> Cmd Msg
updateKifuPage authToken kifuModel =
    let
        position =
            kifuModel.curStep.position
    in
    Cmd.batch
        [ updateBoard ( "shogi", position )
        , Api.requestAsync ApiResponse authToken <|
            Api.KifuRequest <|
                PB.KifuRequest <|
                    PB.RequestGetSamePositions
                        { position = position
                        , steps = 5
                        , excludeKifuIds = [ kifuModel.kifu.kifuId ]
                        }
        ]


apiResponse : Model -> Api.Request -> Api.Response -> ( Model, Cmd Msg )
apiResponse model req res =
    case res of
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

                                kifuModel =
                                    { kifu = r
                                    , curStep =
                                        Maybe.withDefault Kifu.initStep <|
                                            elem r.steps curSeq
                                    , len = List.length r.steps
                                    , samePos = []
                                    }

                                model_ =
                                    { model | kifuModel = kifuModel }

                                authToken =
                                    Maybe.map (\at -> at.token) model.authToken
                            in
                            ( model_
                            , updateKifuPage authToken kifuModel
                            )

                        PB.ResponseGetSamePositions r ->
                            let
                                kifuModel =
                                    model.kifuModel
                            in
                            ( { model | kifuModel = { kifuModel | samePos = r.kifus } }
                            , Cmd.none
                            )

                        _ ->
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

                        km_ =
                            { km
                                | curStep = Maybe.withDefault km.curStep <| elem km.kifu.steps seq
                            }
                    in
                    if km.kifu.kifuId == kifuId then
                        ( { model_ | kifuModel = km_ }
                        , updateKifuPage authToken km_
                        )

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

        OnResize w h ->
            ( { model | windowSize = ( w, h ) }, Cmd.none )

        Logout ->
            ( { model | authToken = Nothing }, removeTokens () )

        _ ->
            --let
            --    _ =
            --        Debug.log (Debug.toString msg) msg
            --in
            ( model, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Browser.Events.onResize OnResize


routeToTitle : Route -> String
routeToTitle route =
    case route of
        Route.Index ->
            "Index"

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
            [ Element.link [] { url = model.loginFormURL, label = Element.text "Signup/Signin" }
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
