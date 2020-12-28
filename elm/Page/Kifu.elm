module Page.Kifu exposing (Model, Msg(..), init, initStep, update, view)

import Element exposing (Element)
import Element.Input as Input
import Html exposing (Html)
import Html.Attributes as Attr
import Proto.Api as PB
import Style


type alias Model =
    { kifu : PB.GetKifuResponse
    , curStep : PB.GetKifuResponse_Step
    , len : Int
    , samePos : List PB.GetSamePositionsResponse_Kifu
    }


initStep : PB.GetKifuResponse_Step
initStep =
    { seq = 0
    , position = ""
    , src = Nothing
    , dst = Nothing
    , piece = PB.Piece_Null
    , finishedStatus = PB.FinishedStatus_NotFinished
    , promoted = False
    , captured = PB.Piece_Null
    , timestampSec = 0
    , thinkingSec = 0
    , notes = []
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
    , curStep = initStep
    , len = 0
    , samePos = []
    }


type Msg
    = UpdateBoard String Int


update : Msg -> Model -> Model
update msg model =
    model


max : Int -> Int -> Int
max a b =
    if a < b then
        b

    else
        a


min : Int -> Int -> Int
min a b =
    if a < b then
        a

    else
        b


control : (Int -> msg) -> Int -> Int -> Element msg
control msg seq len =
    Element.row [ Element.spacing 10 ]
        [ Input.button Style.button
            { onPress = Just <| msg 0
            , label = Element.text "初形"
            }
        , Element.row []
            [ Input.button Style.button
                { onPress = Just <| msg <| max 0 (seq - 1)
                , label = Element.text "前"
                }
            ]
        , Element.row []
            [ Input.button Style.button
                { onPress = Just <| msg <| min (len - 1) (seq + 1)
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


emptyPos : Maybe PB.Pos -> Bool
emptyPos =
    maybe True <| \p -> p.x == 0 || p.y == 0


stepToString :
    { a
        | seq : Int
        , dst : Maybe PB.Pos
        , src : Maybe PB.Pos
        , piece : PB.Piece_Id
        , promoted : Bool
        , finishedStatus : PB.FinishedStatus_Id
    }
    -> String
stepToString step =
    case ( step.finishedStatus == PB.FinishedStatus_NotFinished, emptyPos step.dst, emptyPos step.src ) of
        ( notFin, False, False ) ->
            String.concat
                [ playerSymbol step.seq
                , maybe "" dstToString step.dst
                , pieceToString step.piece
                , if step.promoted then
                    "成"

                  else
                    ""
                , "("
                , maybe "" srcToString step.src
                , ")"
                , if notFin then
                    ""

                  else
                    String.concat
                        [ " ("
                        , finishedToString step.finishedStatus
                        , ")"
                        ]
                ]

        ( notFin, False, True ) ->
            String.concat
                [ playerSymbol step.seq
                , maybe "" dstToString step.dst
                , pieceToString step.piece
                , "打"
                , if notFin then
                    ""

                  else
                    String.concat
                        [ " ("
                        , finishedToString step.finishedStatus
                        , ")"
                        ]
                ]

        ( False, True, _ ) ->
            finishedToString step.finishedStatus

        ( True, True, _ ) ->
            "error data"


secToString : Int -> String
secToString sec =
    let
        minute =
            sec // 60
    in
    String.concat <|
        if minute == 0 then
            [ String.fromInt sec
            , "秒"
            ]

        else
            [ String.fromInt minute
            , "分"
            , String.fromInt sec
            , "秒"
            ]


stepInfo : PB.GetKifuResponse_Step -> Element msg
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


samePosView : List PB.GetSamePositionsResponse_Kifu -> Element msg
samePosView ps =
    Element.column []
        [ Element.text "同一局面"
        , Element.html <|
            Html.ul [] <|
                List.map (\p -> Html.li [] [ Html.text p.kifuId ]) ps
        ]


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    Element.row [ Element.spacing 20, Element.alignTop, Element.width (Element.px 473) ]
        [ Element.column [ Element.spacing 20 ]
            [ Element.el [ Element.width (Element.px 473), Element.height (Element.px 528) ] <|
                Element.html <|
                    Html.canvas [ Attr.id "shogi" ] []
            , control (msg << UpdateBoard model.kifu.kifuId) model.curStep.seq model.len
            , stepInfo model.curStep
            , case model.samePos of
                [] ->
                    Element.none

                samePos ->
                    samePosView model.samePos
            ]
        , gameInfo model
        ]
