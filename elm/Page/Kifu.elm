module Page.Kifu exposing (Model, Msg(..), init, update, view)

import Element exposing (Element)
import Element.Input as Input
import Html exposing (Html)
import Html.Attributes as Attr
import Proto.Api as PB
import Style


type alias Model =
    { kifu : PB.GetKifuResponse
    , curSeq : Int
    , len : Int
    }


init : Model
init =
    { kifu =
        { userId = ""
        , kifuId = ""
        , startTs = 0
        , endTs = 0
        , handicap = ""
        , gameName = ""
        , firstPlayers = []
        , secondPlayers = []
        , otherFields = []
        , sfen = ""
        , createdTs = 0
        , steps = []
        , note = ""
        }
    , curSeq = 0
    , len = 0
    }


type Msg
    = UpdateBoard String Int


update : Msg -> Model -> Model
update msg model =
    model


control : (Int -> msg) -> Int -> Int -> Element msg
control msg seq len =
    Element.row [ Element.spacing 10 ]
        [ Input.button Style.button
            { onPress = Just <| msg 0
            , label = Element.text "初形"
            }
        , Element.row []
            [ Input.button Style.button
                { onPress = Just <| msg (seq - 1)
                , label = Element.text "前"
                }
            ]
        , Element.row []
            [ Input.button Style.button
                { onPress = Just <| msg (seq + 1)
                , label = Element.text "次"
                }
            ]
        , Element.row []
            [ Input.button Style.button
                { onPress = Just <| msg (len - 1)
                , label = Element.text "終局"
                }
            ]
        ]


handicap : String -> String
handicap code =
    case code of
        "NONE" ->
            "平手"

        "DROP_L" ->
            "香落ち"

        "DROP_L_R" ->
            "右香落ち"

        "DROP_B" ->
            "角落ち"

        "DROP_R" ->
            "飛車落ち"

        "DROP_RL" ->
            "飛香落ち"

        "DROP_TWO" ->
            "二枚落ち"

        "DROP_THREE" ->
            "三枚落ち"

        "DROP_FOUR" ->
            "四枚落ち"

        "DROP_FIVE" ->
            "五枚落ち"

        "DROP_FIVE_L" ->
            "左五枚落ち"

        "DROP_SIX" ->
            "六枚落ち"

        "DROP_EIGHT" ->
            "八枚落ち"

        "DROP_TEN" ->
            "十枚落ち"

        _ ->
            "その他"


dlelem : String -> String -> List (Html msg)
dlelem t d =
    [ Html.dt [] [ Html.text t ]
    , Html.dd [] [ Html.text d ]
    ]


gameInfo : Model -> Element msg
gameInfo model =
    let
        ( fstLabel, sndLabel ) =
            if model.kifu.handicap == "NONE" then
                ( "先手", "後手" )

            else
                ( "上手", "下手" )

        names =
            String.concat << List.intersperse ", " << List.map (\p -> p.name)
    in
    Element.el [ Element.alignTop ] <|
        Element.html <|
            Html.dl [ Attr.class "kifuinfo" ] <|
                List.concat <|
                    [ dlelem "棋戦" model.kifu.gameName
                    , dlelem "手割合" <| handicap model.kifu.handicap
                    , dlelem fstLabel <| names model.kifu.firstPlayers
                    , dlelem sndLabel <| names model.kifu.secondPlayers
                    , dlelem "備考" model.kifu.note
                    ]


elem : Int -> List a -> Maybe a
elem n =
    List.head << List.drop n


odd : Int -> Bool
odd i =
    modBy 2 i /= 0


dstToString : PB.Pos -> String
dstToString pos =
    let
        x =
            String.fromInt pos.x

        y =
            case pos.y of
                1 ->
                    "一"

                2 ->
                    "二"

                3 ->
                    "三"

                4 ->
                    "四"

                5 ->
                    "五"

                6 ->
                    "六"

                7 ->
                    "七"

                8 ->
                    "八"

                9 ->
                    "九"

                _ ->
                    "X"
    in
    x ++ y


srcToString : PB.Pos -> String
srcToString pos =
    String.concat <| List.map String.fromInt [ pos.x, pos.y ]


maybe : b -> (a -> b) -> Maybe a -> b
maybe def f =
    Maybe.withDefault def << Maybe.map f


pieceToString : PB.Piece_Id -> String
pieceToString p =
    case p of
        PB.Piece_Gyoku ->
            "玉"

        PB.Piece_Hisha ->
            "飛"

        PB.Piece_Ryu ->
            "竜"

        PB.Piece_Kaku ->
            "角"

        PB.Piece_Uma ->
            "馬"

        PB.Piece_Kin ->
            "金"

        PB.Piece_Gin ->
            "銀"

        PB.Piece_NariGin ->
            "成銀"

        PB.Piece_Kei ->
            "桂"

        PB.Piece_NariKei ->
            "成桂"

        PB.Piece_Kyou ->
            "香"

        PB.Piece_NariKyou ->
            "成香"

        PB.Piece_Fu ->
            "歩"

        PB.Piece_To ->
            "と"

        _ ->
            ""


finishedToString : PB.FinishedStatus_Id -> String
finishedToString finished =
    case finished of
        PB.FinishedStatus_Suspend ->
            "中断"

        PB.FinishedStatus_Surrender ->
            "投了"

        PB.FinishedStatus_Draw ->
            "引き分け"

        PB.FinishedStatus_RepetitionDraw ->
            "千日手"

        PB.FinishedStatus_Checkmate ->
            "詰み"

        PB.FinishedStatus_OverTimeLimit ->
            "時間切れ"

        PB.FinishedStatus_FoulLoss ->
            "反則負け"

        PB.FinishedStatus_FoulWin ->
            "反則勝ち"

        PB.FinishedStatus_NyugyokuWin ->
            "入玉勝ち"

        _ ->
            ""


playerSymbol : Int -> String
playerSymbol seq =
    if odd seq then
        "☗"

    else
        "☖"


stepToString : PB.Step -> String
stepToString step =
    if step.finishedStatus /= PB.FinishedStatus_NotFinished then
        finishedToString step.finishedStatus

    else
        String.concat <|
            playerSymbol step.seq
                :: maybe "" dstToString step.dst
                :: pieceToString step.piece
                :: (case step.src of
                        Just src ->
                            [ if step.promoted then
                                "成"

                              else
                                ""
                            , "("
                            , srcToString src
                            , ")"
                            ]

                        Nothing ->
                            [ "打" ]
                   )


secToString : Int -> String
secToString sec =
    let
        min =
            sec // 60
    in
    String.concat <|
        if min == 0 then
            [ String.fromInt sec
            , "秒"
            ]

        else
            [ String.fromInt min
            , "分"
            , String.fromInt sec
            , "秒"
            ]


stepInfo : PB.Step -> Element msg
stepInfo step =
    Element.column [ Element.spacing 10 ]
        [ if step.seq == 0 then
            Element.text "開始前"

          else
            Element.row [ Element.spacingXY 20 0 ] <|
                List.map Element.text
                    [ String.fromInt step.seq ++ "手目"
                    , stepToString step
                    , secToString step.thinkingSec
                    ]
        , Element.column []
            [ Element.html <|
                Html.ul
                    [ Attr.style "list-style" "none"
                    , Attr.style "padding" "0"
                    , Attr.style "white-space" "normal"
                    ]
                <|
                    List.map
                        (\t ->
                            Html.li
                                [ Attr.style "margin-bottom" "5px"
                                ]
                                [ Html.text t ]
                        )
                        step.notes
            ]
        ]


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    Element.row [ Element.spacing 20, Element.alignTop, Element.width (Element.px 473) ]
        [ Element.column [ Element.spacing 20 ]
            [ Element.el [ Element.width (Element.px 473), Element.height (Element.px 528) ] <|
                Element.html <|
                    Html.canvas [ Attr.id "shogi" ] []
            , control (msg << UpdateBoard model.kifu.kifuId) model.curSeq model.len
            , maybe Element.none stepInfo <| elem model.curSeq model.kifu.steps
            ]
        , gameInfo model
        ]
