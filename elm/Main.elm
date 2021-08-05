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
import Json.Decode as Decoder exposing (Decoder)
import Page.Kifu as Kifu
import Page.MyPage as MyPage
import Page.Upload as Upload
import Proto.Kifu as PB
import Route exposing (Route)
import Style
import Url exposing (Url)


port storeToken : String -> Cmd msg


port storeTokens : ( String, String ) -> Cmd msg


port removeTokens : () -> Cmd msg


port removeTokensReturn : (Int -> msg) -> Sub msg


port updateBoard : ( String, String ) -> Cmd msg


type alias Flags =
    { token : Maybe String
    , refreshToken : Maybe String
    , windowWidth : Int
    , windowHeight : Int
    , authClientId : String
    , authRedirectURL : String
    , logoutRedirectURL : String
    , tokenEndpoint : String
    }


type Msg
    = UrlRequest Browser.UrlRequest
    | UrlChanged Url
    | OnResize Int Int
    | ApiResponse Api.Request Api.Response
    | UploadMsg Upload.Msg
    | KifuMsg Kifu.Msg
    | Logout
    | RemoveAuthTokens
    | AuthTokenMsg (Result Http.Error AuthTokenResponse)
    | RefreshTokenMsg (Result Http.Error RefreshTokenResponse)
    | NOP


type PreviousState
    = PrevNone
    | PrevRequest Api.Request


type alias AuthToken =
    { refreshToken : String
    , token : String
    }


type alias Model =
    { key : Nav.Key
    , route : Route
    , windowSize : ( Int, Int )
    , authToken : Maybe { refreshToken : String, token : String }
    , errorMessage : Maybe String
    , prevState : PreviousState
    , recentKifu : List PB.RecentKifuResponse_Kifu
    , kifuModel : Kifu.Model
    , uploadModel : Upload.Model
    , authClientId : String
    , authRedirectURL : String
    , loginFormURL : String
    , logoutURL : String
    , logoutRedirectURL : String
    , tokenEndpoint : String
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
                , flags.authRedirectURL
                ]

        logoutURL =
            String.concat
                [ "https://kansousenauth.wagahai.info/logout?client_id="
                , flags.authClientId
                , "&redirect_uri="
                , flags.logoutRedirectURL
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
      , authClientId = flags.authClientId
      , authRedirectURL = flags.authRedirectURL
      , loginFormURL = loginFormURL
      , logoutURL = logoutURL
      , logoutRedirectURL = flags.logoutRedirectURL
      , tokenEndpoint = flags.tokenEndpoint
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


type alias AuthTokenResponse =
    { idToken : String
    , accessToken : String
    , refreshToken : String
    , expiresIn : Int
    , tokenType : String
    }


type alias RefreshTokenResponse =
    { idToken : String
    , accessToken : String
    , expiresIn : Int
    , tokenType : String
    }


authTokenDecoder : Decoder AuthTokenResponse
authTokenDecoder =
    Decoder.map5 AuthTokenResponse
        (Decoder.field "id_token" Decoder.string)
        (Decoder.field "access_token" Decoder.string)
        (Decoder.field "refresh_token" Decoder.string)
        (Decoder.field "expires_in" Decoder.int)
        (Decoder.field "token_type" Decoder.string)


refreshTokenDecoder : Decoder RefreshTokenResponse
refreshTokenDecoder =
    Decoder.map4 RefreshTokenResponse
        (Decoder.field "id_token" Decoder.string)
        (Decoder.field "access_token" Decoder.string)
        (Decoder.field "expires_in" Decoder.int)
        (Decoder.field "token_type" Decoder.string)


tokenRequest : Model -> List ( String, String ) -> Decoder a -> (Result Http.Error a -> msg) -> Cmd msg
tokenRequest model params decoder msg =
    let
        kv ( k, v ) =
            String.concat [ k, "=", v ]

        formParams =
            String.join "&" << List.map kv
    in
    Http.post
        { url = model.tokenEndpoint
        , body =
            Http.stringBody "application/x-www-form-urlencoded" <| formParams params
        , expect = Http.expectJson msg decoder
        }


tokenRefreshRequest : Model -> String -> Cmd Msg
tokenRefreshRequest model refreshToken =
    tokenRequest model
        [ ( "grant_type", "refresh_token" )
        , ( "client_id", model.authClientId )
        , ( "refresh_token", refreshToken )
        ]
        refreshTokenDecoder
        RefreshTokenMsg


authorizationRequest : Model -> String -> Cmd Msg
authorizationRequest model code =
    tokenRequest model
        [ ( "grant_type", "authorization_code" )
        , ( "client_id", model.authClientId )
        , ( "redirect_uri", model.authRedirectURL )
        , ( "code", code )
        ]
        authTokenDecoder
        AuthTokenMsg


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
                    , tokenRefreshRequest model t.refreshToken
                    )

                Nothing ->
                    ( { model | prevState = PrevRequest req }
                    , Nav.load model.loginFormURL
                    )

        Err err ->
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


tokenResponse : Model -> Cmd Msg -> ( Model, Cmd Msg )
tokenResponse model store =
    case model.prevState of
        PrevRequest req ->
            ( { model | prevState = PrevNone }
            , Cmd.batch
                [ store
                , Api.request ApiResponse (Maybe.map (\at -> at.token) model.authToken) req
                ]
            )

        PrevNone ->
            ( { model | prevState = PrevNone }
            , Cmd.batch
                [ store
                , Nav.pushUrl model.key (Route.path Route.MyPage)
                ]
            )


httpError : Model -> Route -> Http.Error -> ( Model, Cmd msg )
httpError model route err =
    let
        errMsg =
            case err of
                Http.BadUrl str ->
                    "Bad URL: " ++ str

                Http.Timeout ->
                    "Timeout"

                Http.NetworkError ->
                    "Network Error"

                Http.BadStatus code ->
                    "Bad status: " ++ String.fromInt code

                Http.BadBody str ->
                    "Bad body: " ++ str
    in
    ( { model
        | errorMessage = Just errMsg
      }
    , Nav.pushUrl model.key (Route.path route)
    )


authResponse :
    Model
    -> Result Http.Error a
    -> (a -> Maybe AuthToken)
    -> (a -> Cmd Msg)
    -> ( Model, Cmd Msg )
authResponse model result f store =
    case result of
        Ok res ->
            tokenResponse
                { model | authToken = f res }
                (store res)

        Err err ->
            httpError model Route.Index err


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

                Route.AuthCallback maybeCode ->
                    case maybeCode of
                        Just code ->
                            ( model_
                            , authorizationRequest model code
                            )

                        Nothing ->
                            ( model_, Cmd.none )

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
            ( { model | authToken = Nothing }
            , removeTokens ()
            )

        RemoveAuthTokens ->
            ( model
            , Nav.load model.logoutURL
            )

        AuthTokenMsg result ->
            authResponse model
                result
                (\res ->
                    Just
                        { refreshToken = res.refreshToken
                        , token = res.idToken
                        }
                )
                (\res -> storeTokens ( res.idToken, res.refreshToken ))

        RefreshTokenMsg result ->
            authResponse model
                result
                (\res ->
                    Maybe.map
                        (\at ->
                            { refreshToken = at.refreshToken
                            , token = res.idToken
                            }
                        )
                        model.authToken
                )
                (\res -> storeToken res.idToken)

        NOP ->
            --let
            --    _ =
            --        Debug.log (Debug.toString msg) msg
            --in
            ( model, Cmd.none )


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.batch
        [ Browser.Events.onResize OnResize
        , removeTokensReturn (\_ -> RemoveAuthTokens)
        ]


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

        Route.AuthCallback _ ->
            "Redirect"

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

        Route.AuthCallback _ ->
            Element.none

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
