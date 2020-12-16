module Page.ConfirmForgotPassword exposing (Model, Msg(..), init, update, view)

import Element exposing (Element)
import Element.Input as Input
import Html
import Proto.Api as PB
import Style


type Msg
    = ChangeUsername String
    | ChangeConfirmationCode String
    | ChangePassword String
    | ChangeShowPassword
    | Submit


type alias Model =
    { params : PB.ConfirmForgotPasswordRequest
    , forgotPasswordResponse : Maybe PB.ForgotPasswordResponse
    , showPassword : Bool
    }


init : Model
init =
    { params =
        { username = ""
        , confirmationCode = ""
        , password = ""
        }
    , forgotPasswordResponse = Nothing
    , showPassword = False
    }


update : Msg -> Model -> Model
update msg model =
    let
        ps =
            model.params
    in
    case msg of
        ChangeUsername str ->
            { model | params = { ps | username = str } }

        ChangeConfirmationCode str ->
            { model | params = { ps | confirmationCode = str } }

        ChangePassword str ->
            { model | params = { ps | password = str } }

        ChangeShowPassword ->
            { model | showPassword = not model.showPassword }

        Submit ->
            model


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    let
        form =
            [ Element.html (Html.h1 [] [ Html.text "Sign in" ])
            , Input.username []
                { onChange = msg << ChangeUsername
                , text = model.params.username
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "Username"
                }
            , Input.text []
                { onChange = msg << ChangeConfirmationCode
                , text = model.params.confirmationCode
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "Confirmation code"
                }
            , Input.newPassword []
                { onChange = msg << ChangePassword
                , text = model.params.password
                , placeholder = Nothing
                , label = Input.labelLeft [] <| Element.text "New password"
                , show = model.showPassword
                }
            , Input.checkbox []
                { onChange = msg << (\_ -> ChangeShowPassword)
                , icon = Input.defaultCheckbox
                , checked = model.showPassword
                , label = Input.labelRight [] <| Element.text "Show password"
                }
            , Input.button Style.submitButton
                { onPress = Just (msg Submit)
                , label = Element.text "Confirm"
                }
            ]
    in
    Element.column Style.mainColumn <|
        case model.forgotPasswordResponse of
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
