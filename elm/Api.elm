module Api exposing (Error(..), Request(..), Response(..), errorToString, request, requestAsync)

import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Proto.Api as PB
import Task


type Request
    = KifuRequest PB.KifuRequest


type Response
    = KifuResponse (Result Error PB.KifuResponse)


type Error
    = ErrorResponse (Http.Response String)
    | ErrorJsonDecode Decode.Error
    | ErrorUnauthorized


errorToString : Error -> String
errorToString err =
    case err of
        ErrorResponse res ->
            case res of
                Http.BadUrl_ str ->
                    "BadUrl: " ++ str

                Http.Timeout_ ->
                    "Timeout"

                Http.NetworkError_ ->
                    "NetworkError"

                Http.BadStatus_ meta body ->
                    String.concat
                        [ "BadStatus { status = "
                        , String.fromInt meta.statusCode
                        , ", body = "
                        , body
                        , "}"
                        ]

                Http.GoodStatus_ meta body ->
                    String.concat
                        [ "GoodStatus { status = "
                        , String.fromInt meta.statusCode
                        , ", body = "
                        , body
                        , "}"
                        ]

        ErrorJsonDecode e ->
            Decode.errorToString e

        ErrorUnauthorized ->
            "Unauthorized"


endpoint : String
endpoint =
    "https://kansousenapi.wagahai.info/v1"


headers : List Http.Header
headers =
    []


resultJson : Decoder a -> Http.Response String -> Result Error a
resultJson decoder res =
    case res of
        Http.BadStatus_ meta _ ->
            if meta.statusCode == 401 then
                Err ErrorUnauthorized

            else
                Err <| ErrorResponse res

        Http.GoodStatus_ _ body ->
            case Decode.decodeString decoder body of
                Ok v ->
                    Ok v

                Err err ->
                    Err <| ErrorJsonDecode err

        _ ->
            Err <| ErrorResponse res


request : (Request -> Response -> msg) -> Maybe String -> Request -> Cmd msg
request msg token req =
    case req of
        KifuRequest kifuReq ->
            Http.request
                { method = "POST"
                , headers =
                    case token of
                        Just t ->
                            Http.header "Authorization" t :: headers

                        Nothing ->
                            headers
                , url = endpoint ++ "/kifu"
                , body = Http.jsonBody <| PB.kifuRequestEncoder kifuReq
                , expect =
                    Http.expectStringResponse
                        (msg req << KifuResponse)
                        (resultJson PB.kifuResponseDecoder)
                , timeout = Nothing
                , tracker = Nothing
                }


resolverJson : Decoder a -> Http.Resolver Error a
resolverJson decoder =
    Http.stringResolver <|
        \res ->
            case res of
                Http.BadStatus_ meta _ ->
                    if meta.statusCode == 401 then
                        Err ErrorUnauthorized

                    else
                        Err <| ErrorResponse res

                Http.GoodStatus_ _ body ->
                    case Decode.decodeString decoder body of
                        Ok v ->
                            Ok v

                        Err err ->
                            Err <| ErrorJsonDecode err

                _ ->
                    Err <| ErrorResponse res


requestAsync : (Request -> Response -> msg) -> Maybe String -> Request -> Cmd msg
requestAsync msg token req =
    Cmd.map (msg req) <|
        case req of
            KifuRequest kifuReq ->
                Task.attempt KifuResponse <|
                    Http.task
                        { method = "POST"
                        , headers =
                            case token of
                                Just t ->
                                    Http.header "Authorization" t :: headers

                                Nothing ->
                                    headers
                        , url = endpoint ++ "/kifu"
                        , body = Http.jsonBody <| PB.kifuRequestEncoder kifuReq
                        , resolver = resolverJson PB.kifuResponseDecoder
                        , timeout = Nothing
                        }
