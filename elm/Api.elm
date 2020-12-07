module Api exposing (Error(..), Request(..), Response(..), errorToString, request)

import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Proto.Api as PB


type Request
    = AuthRequest PB.AuthRequest
    | HelloRequest


type Response
    = AuthResponse (Result Error PB.AuthResponse)
    | HelloResponse (Result Error PB.HelloResponse)


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
    "https://kifuapi.wagahai.info/v1"


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
        AuthRequest authreq ->
            Http.request
                { method = "POST"
                , headers = headers
                , url = endpoint ++ "/auth"
                , body = Http.jsonBody <| PB.authRequestEncoder authreq
                , expect =
                    Http.expectStringResponse
                        (msg req << AuthResponse)
                        (resultJson PB.authResponseDecoder)
                , timeout = Nothing
                , tracker = Nothing
                }

        HelloRequest ->
            Http.request
                { method = "POST"
                , headers =
                    case token of
                        Just t ->
                            Http.header "Authorization" t :: headers

                        Nothing ->
                            headers
                , url = endpoint ++ "/hello"
                , body = Http.stringBody "application/json" "{}"
                , expect =
                    Http.expectStringResponse
                        (msg req << HelloResponse)
                        (resultJson PB.helloResponseDecoder)
                , timeout = Nothing
                , tracker = Nothing
                }
