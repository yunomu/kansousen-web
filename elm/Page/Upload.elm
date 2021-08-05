module Page.Upload exposing (..)

import Element exposing (Element)
import Element.Input as Input
import Html
import Proto.Kifu as Api
import Style


type Msg
    = Submit
    | ChangeKifu String
    | ChangeRepeat


type alias Model =
    { request : Api.PostKifuRequest
    , repeat : Bool
    }


init : Bool -> Model
init repeat =
    { request =
        { payload = ""
        , format = "KIF"
        , encoding = "UTF-8"
        }
    , repeat = repeat
    }


update : Msg -> Model -> Model
update msg model =
    case msg of
        Submit ->
            model

        ChangeKifu str ->
            let
                req =
                    model.request
            in
            { model | request = { req | payload = str } }

        ChangeRepeat ->
            { model | repeat = not model.repeat }


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    Element.column Style.mainColumn
        [ Element.html (Html.h1 [] [ Html.text "Upload" ])
        , Input.multiline Style.kifuField
            { onChange = msg << ChangeKifu
            , text = model.request.payload
            , placeholder = Nothing
            , label = Input.labelAbove [] (Element.text "棋譜")
            , spellcheck = False
            }
        , Input.checkbox []
            { onChange = \_ -> msg ChangeRepeat
            , icon = Input.defaultCheckbox
            , checked = model.repeat
            , label = Input.labelRight [] <| Element.text "続けて入力する"
            }
        , Input.button Style.submitButton
            { onPress = Just (msg Submit)
            , label = Element.text "Submit"
            }
        ]
