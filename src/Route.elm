module Route exposing (Route(..), fromUrl, path)

import Url exposing (Url)
import Url.Builder as UrlBuilder
import Url.Parser as P exposing ((</>), (<?>), Parser, s)
import Url.Parser.Query as Query


type Route
    = Index
    | MyPage
    | Upload
    | Kifu String Int
    | AuthCallback (Maybe String)
    | NotFound


parser : Parser (Route -> a) a
parser =
    P.oneOf
        [ P.map Index P.top
        , P.map Index <| s "index.html"
        , P.map MyPage <| s "my"
        , P.map Upload <| s "upload"
        , P.map Kifu <| s "kifu" </> P.string </> P.int
        , P.map (\id -> Kifu id 0) <| s "kifu" </> P.string
        , P.map (\id -> Kifu id 0) <| s "kifu" </> P.string </> s ""
        , P.map AuthCallback <| s "callback" <?> Query.string "code"
        ]


path : Route -> String
path route =
    case route of
        Index ->
            UrlBuilder.absolute [] []

        MyPage ->
            UrlBuilder.absolute [ "my" ] []

        Upload ->
            UrlBuilder.absolute [ "upload" ] []

        Kifu kifuId seq ->
            UrlBuilder.absolute [ "kifu", kifuId, String.fromInt seq ] []

        _ ->
            UrlBuilder.absolute [] []


fromUrl : Url -> Route
fromUrl url =
    Maybe.withDefault NotFound (P.parse parser url)
