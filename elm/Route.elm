module Route exposing (Route(..), fromUrl, path)

import Url exposing (Url)
import Url.Builder as UrlBuilder
import Url.Parser as UrlParser exposing ((</>), Parser, s)


type Route
    = Index
    | SignUp
    | ConfirmSignUp
    | ResendConfirm
    | ForgotPassword
    | ConfirmForgotPassword
    | SignIn
    | MyPage
    | Upload
    | NotFound


parser : Parser (Route -> a) a
parser =
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
        , UrlParser.map Upload <| s "upload"
        ]


path : Route -> String
path route =
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

        Upload ->
            UrlBuilder.absolute [ "upload" ] []

        NotFound ->
            UrlBuilder.absolute [] []


fromUrl : Url -> Route
fromUrl url =
    Maybe.withDefault NotFound (UrlParser.parse parser url)
