module Page.ResendConfirm exposing (Model, Msg(..), init, update, view)

import Element exposing (Element)
import Element.Input as Input
import Proto.Api as PB


type Msg
    = ChangeName String
    | Submit


type alias Model =
    PB.ResendConfirmationCodeRequest


init : Model
init =
    { username = ""
    }


update : Msg -> Model -> Model
update msg model =
    case msg of
        ChangeName str ->
            { model | username = str }

        Submit ->
            model


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    Element.column []
        [ Input.username []
            { onChange = msg << ChangeName
            , text = model.username
            , placeholder = Nothing
            , label = Input.labelLeft [] <| Element.text "Username"
            }
        , Input.button []
            { onPress = Just (msg Submit)
            , label = Element.text "Submit"
            }
        ]
