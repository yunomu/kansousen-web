module Api exposing
    ( Request(..)
    , Response(..)
    , request
    , requestAsync
    )

import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode
import Proto.Kifu as PB
import Task exposing (Task)


type Request
    = GetKifuRequest PB.GetKifuRequest
    | PostKifuRequest PB.PostKifuRequest
    | DeleteKifuRequest PB.DeleteKifuRequest
    | RecentKifuRequest PB.RecentKifuRequest
    | GetSamePositionsRequest PB.GetSamePositionsRequest


requestEncoder : Request -> Encode.Value
requestEncoder req =
    case req of
        GetKifuRequest r ->
            PB.getKifuRequestEncoder r

        PostKifuRequest r ->
            PB.postKifuRequestEncoder r

        DeleteKifuRequest r ->
            PB.deleteKifuRequestEncoder r

        RecentKifuRequest r ->
            PB.recentKifuRequestEncoder r

        GetSamePositionsRequest r ->
            PB.getSamePositionsRequestEncoder r


requestPath : Request -> String
requestPath req =
    case req of
        GetKifuRequest _ ->
            "/get-kifu"

        PostKifuRequest _ ->
            "/post-kifu"

        DeleteKifuRequest _ ->
            "/delete-kifu"

        RecentKifuRequest _ ->
            "/recent-kifu"

        GetSamePositionsRequest _ ->
            "/same-positions"


type Response
    = GetKifuResponse PB.GetKifuResponse
    | PostKifuResponse PB.PostKifuResponse
    | DeleteKifuResponse PB.DeleteKifuResponse
    | RecentKifuResponse PB.RecentKifuResponse
    | GetSamePositionsResponse PB.GetSamePositionsResponse
    | ErrorJsonDecode Decode.Error
    | Unauthenticated Request


responseDecoder : Request -> Decoder Response
responseDecoder req =
    case req of
        GetKifuRequest _ ->
            Decode.map GetKifuResponse PB.getKifuResponseDecoder

        PostKifuRequest _ ->
            Decode.map PostKifuResponse PB.postKifuResponseDecoder

        DeleteKifuRequest _ ->
            Decode.map DeleteKifuResponse PB.deleteKifuResponseDecoder

        RecentKifuRequest _ ->
            Decode.map RecentKifuResponse PB.recentKifuResponseDecoder

        GetSamePositionsRequest _ ->
            Decode.map GetSamePositionsResponse PB.getSamePositionsResponseDecoder


endpoint : String
endpoint =
    "https://kansousenapi.wagahai.info/v1"


headers : List Http.Header
headers =
    []


jsonResponse : Request -> Http.Response String -> Result Http.Error Response
jsonResponse req res =
    case res of
        Http.GoodStatus_ _ body ->
            case Decode.decodeString (responseDecoder req) body of
                Ok v ->
                    Ok v

                Err err ->
                    Ok <| ErrorJsonDecode err

        Http.BadUrl_ url ->
            Err <| Http.BadUrl url

        Http.Timeout_ ->
            Err Http.Timeout

        Http.NetworkError_ ->
            Err Http.Timeout

        Http.BadStatus_ meta _ ->
            if meta.statusCode == 401 then
                Ok <| Unauthenticated req

            else
                Err <| Http.BadStatus meta.statusCode


request : (Request -> Result Http.Error Response -> msg) -> Maybe String -> Request -> Cmd msg
request msg token req =
    Http.request
        { method = "POST"
        , headers =
            case token of
                Just t ->
                    Http.header "Authorization" t :: headers

                Nothing ->
                    headers
        , url = endpoint ++ requestPath req
        , body = Http.jsonBody <| requestEncoder req
        , expect =
            Http.expectStringResponse
                (msg req)
                (jsonResponse req)
        , timeout = Nothing
        , tracker = Nothing
        }


requestAsync : Maybe String -> Request -> Task Http.Error Response
requestAsync token req =
    Http.task
        { method = "POST"
        , headers =
            case token of
                Just t ->
                    Http.header "Authorization" t :: headers

                Nothing ->
                    headers
        , url = endpoint ++ requestPath req
        , body = Http.jsonBody <| requestEncoder req
        , resolver =
            Http.stringResolver <| jsonResponse req
        , timeout = Nothing
        }
