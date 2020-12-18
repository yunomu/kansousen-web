module Route exposing (Route(..), fromUrl, path)

import Url exposing (Url)
import Url.Builder as UrlBuilder
import Url.Parser as P exposing ((</>), Parser, s)


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
    | Kifu String Int
    | NotFound


parser : Parser (Route -> a) a
parser =
    P.oneOf
        [ P.map Index P.top
        , P.map Index <| s "index.html"
        , P.map SignUp <| s "signup"
        , P.map ConfirmSignUp <| s "confirm_signup"
        , P.map ResendConfirm <| s "resend_confirm"
        , P.map SignIn <| s "signin"
        , P.map ForgotPassword <| s "forgot_password"
        , P.map ConfirmForgotPassword <| s "confirm_forgot_password"
        , P.map MyPage <| s "my"
        , P.map Upload <| s "upload"
        , P.map Kifu <| s "kifu" </> P.string </> P.int
        , P.map (\id -> Kifu id 0) <| s "kifu" </> P.string
        , P.map (\id -> Kifu id 0) <| s "kifu" </> P.string </> s ""
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

        Kifu kifuId seq ->
            UrlBuilder.absolute [ "kifu", kifuId, String.fromInt seq ] []

        NotFound ->
            UrlBuilder.absolute [] []


fromUrl : Url -> Route
fromUrl url =
    Maybe.withDefault NotFound (P.parse parser url)
