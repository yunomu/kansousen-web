module Page.ConfirmSignUp exposing (Model, Msg(..), init, update, view)

import Element exposing (Element)
import Element.Input as Input
import Proto.Api as PB
import Route
import Style


type Msg
    = ChangeName String
    | ChangeCode String
    | Submit


type alias Model =
    { params : PB.ConfirmSignUpRequest
    , signUpResponse : Maybe PB.SignUpResponse
    }


init : Model
init =
    { params =
        { username = ""
        , confirmationCode = ""
        }
    , signUpResponse = Nothing
    }


update : Msg -> Model -> Model
update msg model =
    let
        params =
            model.params
    in
    case msg of
        ChangeName str ->
            { model | params = { params | username = str } }

        ChangeCode str ->
            { model | params = { params | confirmationCode = str } }

        Submit ->
            model


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    let
        form =
            [ Input.username []
                { onChange = msg << ChangeName
                , text = model.params.username
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "Username"
                }
            , Input.text []
                { onChange = msg << ChangeCode
                , text = model.params.confirmationCode
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "Confirmation code"
                }
            , Input.button Style.submitButton
                { onPress = Just (msg Submit)
                , label = Element.text "Submit"
                }
            , Element.link []
                { url = Route.path Route.ResendConfirm
                , label = Element.text "Resend confirmation code"
                }
            ]
    in
    Element.column Style.mainColumn <|
        case model.signUpResponse of
            Just res ->
                (Element.el [] <|
                    Element.text <|
                        "I sent the confirmation code to `"
                            ++ res.codeDeliveryDestination
                            ++ "` via  "
                            ++ res.codeDeliveryType
                            ++ "."
                )
                    :: form

            Nothing ->
                form
