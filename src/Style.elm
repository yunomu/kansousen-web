module Style exposing (..)

import Element exposing (Attribute, Color, Element, px)
import Element.Background as Background
import Element.Border as Border
import Element.Font as Font


kifuField : List (Attribute msg)
kifuField =
    [ Font.size 15
    , Element.width (px 500)
    , Element.height (px 700)
    ]


submitFontColor : Color
submitFontColor =
    Element.rgb 1 1 1


submitBackgroundColor : Color
submitBackgroundColor =
    Element.rgb255 0 0 255


submitButton : List (Attribute msg)
submitButton =
    [ Font.color submitFontColor
    , Background.color submitBackgroundColor
    , Border.rounded 3
    , Element.padding 3
    ]


button : List (Attribute msg)
button =
    [ Border.rounded 3
    , Border.color (Element.rgb 0 0 0)
    , Border.solid
    , Border.width 2
    , Element.padding 3
    ]


border : Element msg
border =
    Element.el
        [ Element.width (px 1)
        , Element.height Element.fill
        , Background.color (Element.rgb 0 0 0)
        ]
        Element.none


mainColumn : List (Attribute msg)
mainColumn =
    [ Element.spacing 5
    ]
