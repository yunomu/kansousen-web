module Page.SignUp exposing (Model, Msg(..), init, update, view)

import Element exposing (Element)
import Element.Input as Input
import Html
import Proto.Api as PB
import Style


type Msg
    = ChangeUsername String
    | ChangePassword String
    | ChangeShowPassword
    | Submit


type alias Model =
    { params : PB.SignUpRequest
    , showPassword : Bool
    }


init : Model
init =
    { params =
        { username = ""
        , password = ""
        }
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

        ChangePassword str ->
            { model | params = { ps | password = str } }

        ChangeShowPassword ->
            { model | showPassword = not model.showPassword }

        Submit ->
            model


view : (Msg -> msg) -> Model -> Element msg
view msg model =
    Element.column Style.mainColumn
        [ Element.html (Html.h1 [] [ Html.text "Sign up" ])
        , Input.email []
            { onChange = msg << ChangeUsername
            , text = model.params.username
            , placeholder = Nothing
            , label = Input.labelLeft [] <| Element.text "Username"
            }
        , Input.newPassword []
            { onChange = msg << ChangePassword
            , text = model.params.password
            , placeholder = Nothing
            , label = Input.labelLeft [] <| Element.text "Password"
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
            , label = Element.text "SignUp"
            }
        ]
