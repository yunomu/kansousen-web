module Page.MyPage exposing (view)

import Element exposing (Element)
import Proto.Api as PB
import Route


type alias Kifu =
    PB.RecentKifuResponse_Kifu


url : Kifu -> String
url kifu =
    "/kifu/" ++ kifu.kifuId


label : Kifu -> Element msg
label kifu =
    let
        t =
            String.concat
                [ kifu.gameName
                , ": "
                , kifu.firstPlayer
                , " vs "
                , kifu.secondPlayer
                ]
    in
    Element.text <|
        case t of
            "" ->
                "..."

            _ ->
                t


view : List Kifu -> Element msg
view kifus =
    Element.column Style.mainColumn
        [ Element.link []
            { url = Route.path Route.Upload
            , label = Element.text "アップロード"
            }
        , Element.text "最近の棋譜"
        , Element.column [] <|
            List.map
                (\kifu ->
                    Element.link []
                        { url = url kifu
                        , label = label kifu
                        }
                )
                kifus
        ]
