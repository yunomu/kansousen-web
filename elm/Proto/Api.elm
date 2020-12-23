module Proto.Api exposing (..)

-- DO NOT EDIT
-- AUTOGENERATED BY THE ELM PROTOCOL BUFFER COMPILER
-- https://github.com/tiziano88/elm-protobuf
-- source file: proto/api.proto

import Protobuf exposing (..)

import Json.Decode as JD
import Json.Encode as JE


uselessDeclarationToPreventErrorDueToEmptyOutputFile = 42


type alias SignUpRequest =
    { username : String -- 1
    , email : String -- 2
    , password : String -- 3
    }


signUpRequestDecoder : JD.Decoder SignUpRequest
signUpRequestDecoder =
    JD.lazy <| \_ -> decode SignUpRequest
        |> required "username" JD.string ""
        |> required "email" JD.string ""
        |> required "password" JD.string ""


signUpRequestEncoder : SignUpRequest -> JE.Value
signUpRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        , (requiredFieldEncoder "email" JE.string "" v.email)
        , (requiredFieldEncoder "password" JE.string "" v.password)
        ]


type alias ConfirmSignUpRequest =
    { username : String -- 1
    , confirmationCode : String -- 2
    }


confirmSignUpRequestDecoder : JD.Decoder ConfirmSignUpRequest
confirmSignUpRequestDecoder =
    JD.lazy <| \_ -> decode ConfirmSignUpRequest
        |> required "username" JD.string ""
        |> required "confirmationCode" JD.string ""


confirmSignUpRequestEncoder : ConfirmSignUpRequest -> JE.Value
confirmSignUpRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        , (requiredFieldEncoder "confirmationCode" JE.string "" v.confirmationCode)
        ]


type alias ResendConfirmationCodeRequest =
    { username : String -- 1
    }


resendConfirmationCodeRequestDecoder : JD.Decoder ResendConfirmationCodeRequest
resendConfirmationCodeRequestDecoder =
    JD.lazy <| \_ -> decode ResendConfirmationCodeRequest
        |> required "username" JD.string ""


resendConfirmationCodeRequestEncoder : ResendConfirmationCodeRequest -> JE.Value
resendConfirmationCodeRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        ]


type alias ForgotPasswordRequest =
    { username : String -- 1
    }


forgotPasswordRequestDecoder : JD.Decoder ForgotPasswordRequest
forgotPasswordRequestDecoder =
    JD.lazy <| \_ -> decode ForgotPasswordRequest
        |> required "username" JD.string ""


forgotPasswordRequestEncoder : ForgotPasswordRequest -> JE.Value
forgotPasswordRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        ]


type alias ConfirmForgotPasswordRequest =
    { username : String -- 1
    , password : String -- 2
    , confirmationCode : String -- 3
    }


confirmForgotPasswordRequestDecoder : JD.Decoder ConfirmForgotPasswordRequest
confirmForgotPasswordRequestDecoder =
    JD.lazy <| \_ -> decode ConfirmForgotPasswordRequest
        |> required "username" JD.string ""
        |> required "password" JD.string ""
        |> required "confirmationCode" JD.string ""


confirmForgotPasswordRequestEncoder : ConfirmForgotPasswordRequest -> JE.Value
confirmForgotPasswordRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        , (requiredFieldEncoder "password" JE.string "" v.password)
        , (requiredFieldEncoder "confirmationCode" JE.string "" v.confirmationCode)
        ]


type alias SignInRequest =
    { username : String -- 1
    , password : String -- 2
    }


signInRequestDecoder : JD.Decoder SignInRequest
signInRequestDecoder =
    JD.lazy <| \_ -> decode SignInRequest
        |> required "username" JD.string ""
        |> required "password" JD.string ""


signInRequestEncoder : SignInRequest -> JE.Value
signInRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "username" JE.string "" v.username)
        , (requiredFieldEncoder "password" JE.string "" v.password)
        ]


type alias TokenRefreshRequest =
    { refreshToken : String -- 1
    }


tokenRefreshRequestDecoder : JD.Decoder TokenRefreshRequest
tokenRefreshRequestDecoder =
    JD.lazy <| \_ -> decode TokenRefreshRequest
        |> required "refreshToken" JD.string ""


tokenRefreshRequestEncoder : TokenRefreshRequest -> JE.Value
tokenRefreshRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "refreshToken" JE.string "" v.refreshToken)
        ]


type alias AuthRequest =
    { authRequestSelect : AuthRequestSelect
    }


type AuthRequestSelect
    = AuthRequestSelectUnspecified
    | RequestSignUp SignUpRequest
    | RequestConfirmSignUp ConfirmSignUpRequest
    | RequestResendConfirmationCode ResendConfirmationCodeRequest
    | RequestForgotPassword ForgotPasswordRequest
    | RequestConfirmForgotPassword ConfirmForgotPasswordRequest
    | RequestSignIn SignInRequest
    | RequestTokenRefresh TokenRefreshRequest


authRequestSelectDecoder : JD.Decoder AuthRequestSelect
authRequestSelectDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map RequestSignUp (JD.field "requestSignUp" signUpRequestDecoder)
        , JD.map RequestConfirmSignUp (JD.field "requestConfirmSignUp" confirmSignUpRequestDecoder)
        , JD.map RequestResendConfirmationCode (JD.field "requestResendConfirmationCode" resendConfirmationCodeRequestDecoder)
        , JD.map RequestForgotPassword (JD.field "requestForgotPassword" forgotPasswordRequestDecoder)
        , JD.map RequestConfirmForgotPassword (JD.field "requestConfirmForgotPassword" confirmForgotPasswordRequestDecoder)
        , JD.map RequestSignIn (JD.field "requestSignIn" signInRequestDecoder)
        , JD.map RequestTokenRefresh (JD.field "requestTokenRefresh" tokenRefreshRequestDecoder)
        , JD.succeed AuthRequestSelectUnspecified
        ]


authRequestSelectEncoder : AuthRequestSelect -> Maybe ( String, JE.Value )
authRequestSelectEncoder v =
    case v of
        AuthRequestSelectUnspecified ->
            Nothing
        RequestSignUp x ->
            Just ( "requestSignUp", signUpRequestEncoder x )
        RequestConfirmSignUp x ->
            Just ( "requestConfirmSignUp", confirmSignUpRequestEncoder x )
        RequestResendConfirmationCode x ->
            Just ( "requestResendConfirmationCode", resendConfirmationCodeRequestEncoder x )
        RequestForgotPassword x ->
            Just ( "requestForgotPassword", forgotPasswordRequestEncoder x )
        RequestConfirmForgotPassword x ->
            Just ( "requestConfirmForgotPassword", confirmForgotPasswordRequestEncoder x )
        RequestSignIn x ->
            Just ( "requestSignIn", signInRequestEncoder x )
        RequestTokenRefresh x ->
            Just ( "requestTokenRefresh", tokenRefreshRequestEncoder x )


authRequestDecoder : JD.Decoder AuthRequest
authRequestDecoder =
    JD.lazy <| \_ -> decode AuthRequest
        |> field authRequestSelectDecoder


authRequestEncoder : AuthRequest -> JE.Value
authRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (authRequestSelectEncoder v.authRequestSelect)
        ]


type alias SignUpResponse =
    { codeDeliveryType : String -- 1
    , codeDeliveryDestination : String -- 2
    }


signUpResponseDecoder : JD.Decoder SignUpResponse
signUpResponseDecoder =
    JD.lazy <| \_ -> decode SignUpResponse
        |> required "codeDeliveryType" JD.string ""
        |> required "codeDeliveryDestination" JD.string ""


signUpResponseEncoder : SignUpResponse -> JE.Value
signUpResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "codeDeliveryType" JE.string "" v.codeDeliveryType)
        , (requiredFieldEncoder "codeDeliveryDestination" JE.string "" v.codeDeliveryDestination)
        ]


type alias ConfirmSignUpResponse =
    {
    }


confirmSignUpResponseDecoder : JD.Decoder ConfirmSignUpResponse
confirmSignUpResponseDecoder =
    JD.lazy <| \_ -> decode ConfirmSignUpResponse


confirmSignUpResponseEncoder : ConfirmSignUpResponse -> JE.Value
confirmSignUpResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


type alias ForgotPasswordResponse =
    { codeDeliveryType : String -- 1
    , codeDeliveryDestination : String -- 2
    }


forgotPasswordResponseDecoder : JD.Decoder ForgotPasswordResponse
forgotPasswordResponseDecoder =
    JD.lazy <| \_ -> decode ForgotPasswordResponse
        |> required "codeDeliveryType" JD.string ""
        |> required "codeDeliveryDestination" JD.string ""


forgotPasswordResponseEncoder : ForgotPasswordResponse -> JE.Value
forgotPasswordResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "codeDeliveryType" JE.string "" v.codeDeliveryType)
        , (requiredFieldEncoder "codeDeliveryDestination" JE.string "" v.codeDeliveryDestination)
        ]


type alias ConfirmForgotPasswordResponse =
    {
    }


confirmForgotPasswordResponseDecoder : JD.Decoder ConfirmForgotPasswordResponse
confirmForgotPasswordResponseDecoder =
    JD.lazy <| \_ -> decode ConfirmForgotPasswordResponse


confirmForgotPasswordResponseEncoder : ConfirmForgotPasswordResponse -> JE.Value
confirmForgotPasswordResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


type alias SignInResponse =
    { token : String -- 1
    , refreshToken : String -- 2
    }


signInResponseDecoder : JD.Decoder SignInResponse
signInResponseDecoder =
    JD.lazy <| \_ -> decode SignInResponse
        |> required "token" JD.string ""
        |> required "refreshToken" JD.string ""


signInResponseEncoder : SignInResponse -> JE.Value
signInResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "token" JE.string "" v.token)
        , (requiredFieldEncoder "refreshToken" JE.string "" v.refreshToken)
        ]


type alias TokenRefreshResponse =
    { token : String -- 1
    }


tokenRefreshResponseDecoder : JD.Decoder TokenRefreshResponse
tokenRefreshResponseDecoder =
    JD.lazy <| \_ -> decode TokenRefreshResponse
        |> required "token" JD.string ""


tokenRefreshResponseEncoder : TokenRefreshResponse -> JE.Value
tokenRefreshResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "token" JE.string "" v.token)
        ]


type alias AuthResponse =
    { authResponseSelect : AuthResponseSelect
    }


type AuthResponseSelect
    = AuthResponseSelectUnspecified
    | ResponseSignUp SignUpResponse
    | ResponseConfirmSignUp ConfirmSignUpResponse
    | ResponseForgotPassword ForgotPasswordResponse
    | ResponseConfirmForgotPassword ConfirmForgotPasswordResponse
    | ResponseSignIn SignInResponse
    | ResponseTokenRefresh TokenRefreshResponse


authResponseSelectDecoder : JD.Decoder AuthResponseSelect
authResponseSelectDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map ResponseSignUp (JD.field "responseSignUp" signUpResponseDecoder)
        , JD.map ResponseConfirmSignUp (JD.field "responseConfirmSignUp" confirmSignUpResponseDecoder)
        , JD.map ResponseForgotPassword (JD.field "responseForgotPassword" forgotPasswordResponseDecoder)
        , JD.map ResponseConfirmForgotPassword (JD.field "responseConfirmForgotPassword" confirmForgotPasswordResponseDecoder)
        , JD.map ResponseSignIn (JD.field "responseSignIn" signInResponseDecoder)
        , JD.map ResponseTokenRefresh (JD.field "responseTokenRefresh" tokenRefreshResponseDecoder)
        , JD.succeed AuthResponseSelectUnspecified
        ]


authResponseSelectEncoder : AuthResponseSelect -> Maybe ( String, JE.Value )
authResponseSelectEncoder v =
    case v of
        AuthResponseSelectUnspecified ->
            Nothing
        ResponseSignUp x ->
            Just ( "responseSignUp", signUpResponseEncoder x )
        ResponseConfirmSignUp x ->
            Just ( "responseConfirmSignUp", confirmSignUpResponseEncoder x )
        ResponseForgotPassword x ->
            Just ( "responseForgotPassword", forgotPasswordResponseEncoder x )
        ResponseConfirmForgotPassword x ->
            Just ( "responseConfirmForgotPassword", confirmForgotPasswordResponseEncoder x )
        ResponseSignIn x ->
            Just ( "responseSignIn", signInResponseEncoder x )
        ResponseTokenRefresh x ->
            Just ( "responseTokenRefresh", tokenRefreshResponseEncoder x )


authResponseDecoder : JD.Decoder AuthResponse
authResponseDecoder =
    JD.lazy <| \_ -> decode AuthResponse
        |> field authResponseSelectDecoder


authResponseEncoder : AuthResponse -> JE.Value
authResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (authResponseSelectEncoder v.authResponseSelect)
        ]


type alias RecentKifuRequest =
    { limit : Int -- 1
    }


recentKifuRequestDecoder : JD.Decoder RecentKifuRequest
recentKifuRequestDecoder =
    JD.lazy <| \_ -> decode RecentKifuRequest
        |> required "limit" intDecoder 0


recentKifuRequestEncoder : RecentKifuRequest -> JE.Value
recentKifuRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "limit" JE.int 0 v.limit)
        ]


type alias RecentKifuResponse =
    { kifus : List RecentKifuResponse_Kifu -- 1
    }


recentKifuResponseDecoder : JD.Decoder RecentKifuResponse
recentKifuResponseDecoder =
    JD.lazy <| \_ -> decode RecentKifuResponse
        |> repeated "kifus" recentKifuResponse_KifuDecoder


recentKifuResponseEncoder : RecentKifuResponse -> JE.Value
recentKifuResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (repeatedFieldEncoder "kifus" recentKifuResponse_KifuEncoder v.kifus)
        ]


type alias RecentKifuResponse_Kifu =
    { userId : String -- 1
    , kifuId : String -- 2
    , startTs : Int -- 3
    , handicap : String -- 4
    , gameName : String -- 5
    , firstPlayer : String -- 6
    , secondPlayer : String -- 7
    , note : String -- 8
    }


recentKifuResponse_KifuDecoder : JD.Decoder RecentKifuResponse_Kifu
recentKifuResponse_KifuDecoder =
    JD.lazy <| \_ -> decode RecentKifuResponse_Kifu
        |> required "userId" JD.string ""
        |> required "kifuId" JD.string ""
        |> required "startTs" intDecoder 0
        |> required "handicap" JD.string ""
        |> required "gameName" JD.string ""
        |> required "firstPlayer" JD.string ""
        |> required "secondPlayer" JD.string ""
        |> required "note" JD.string ""


recentKifuResponse_KifuEncoder : RecentKifuResponse_Kifu -> JE.Value
recentKifuResponse_KifuEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "userId" JE.string "" v.userId)
        , (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        , (requiredFieldEncoder "startTs" numericStringEncoder 0 v.startTs)
        , (requiredFieldEncoder "handicap" JE.string "" v.handicap)
        , (requiredFieldEncoder "gameName" JE.string "" v.gameName)
        , (requiredFieldEncoder "firstPlayer" JE.string "" v.firstPlayer)
        , (requiredFieldEncoder "secondPlayer" JE.string "" v.secondPlayer)
        , (requiredFieldEncoder "note" JE.string "" v.note)
        ]


type alias PostKifuRequest =
    { payload : String -- 1
    , format : String -- 2
    , encoding : String -- 3
    }


postKifuRequestDecoder : JD.Decoder PostKifuRequest
postKifuRequestDecoder =
    JD.lazy <| \_ -> decode PostKifuRequest
        |> required "payload" JD.string ""
        |> required "format" JD.string ""
        |> required "encoding" JD.string ""


postKifuRequestEncoder : PostKifuRequest -> JE.Value
postKifuRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "payload" JE.string "" v.payload)
        , (requiredFieldEncoder "format" JE.string "" v.format)
        , (requiredFieldEncoder "encoding" JE.string "" v.encoding)
        ]


type alias PostKifuResponse =
    { kifuId : String -- 1
    , duplicated : List PostKifuResponse_Kifu -- 2
    }


postKifuResponseDecoder : JD.Decoder PostKifuResponse
postKifuResponseDecoder =
    JD.lazy <| \_ -> decode PostKifuResponse
        |> required "kifuId" JD.string ""
        |> repeated "duplicated" postKifuResponse_KifuDecoder


postKifuResponseEncoder : PostKifuResponse -> JE.Value
postKifuResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        , (repeatedFieldEncoder "duplicated" postKifuResponse_KifuEncoder v.duplicated)
        ]


type alias PostKifuResponse_Kifu =
    { userId : String -- 1
    , kifuId : String -- 2
    }


postKifuResponse_KifuDecoder : JD.Decoder PostKifuResponse_Kifu
postKifuResponse_KifuDecoder =
    JD.lazy <| \_ -> decode PostKifuResponse_Kifu
        |> required "userId" JD.string ""
        |> required "kifuId" JD.string ""


postKifuResponse_KifuEncoder : PostKifuResponse_Kifu -> JE.Value
postKifuResponse_KifuEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "userId" JE.string "" v.userId)
        , (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        ]


type alias DeleteKifuRequest =
    { kifuId : String -- 1
    }


deleteKifuRequestDecoder : JD.Decoder DeleteKifuRequest
deleteKifuRequestDecoder =
    JD.lazy <| \_ -> decode DeleteKifuRequest
        |> required "kifuId" JD.string ""


deleteKifuRequestEncoder : DeleteKifuRequest -> JE.Value
deleteKifuRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        ]


type alias DeleteKifuResponse =
    {
    }


deleteKifuResponseDecoder : JD.Decoder DeleteKifuResponse
deleteKifuResponseDecoder =
    JD.lazy <| \_ -> decode DeleteKifuResponse


deleteKifuResponseEncoder : DeleteKifuResponse -> JE.Value
deleteKifuResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


type alias GetKifuRequest =
    { kifuId : String -- 1
    }


getKifuRequestDecoder : JD.Decoder GetKifuRequest
getKifuRequestDecoder =
    JD.lazy <| \_ -> decode GetKifuRequest
        |> required "kifuId" JD.string ""


getKifuRequestEncoder : GetKifuRequest -> JE.Value
getKifuRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        ]


type alias Pos =
    { x : Int -- 1
    , y : Int -- 2
    }


posDecoder : JD.Decoder Pos
posDecoder =
    JD.lazy <| \_ -> decode Pos
        |> required "x" intDecoder 0
        |> required "y" intDecoder 0


posEncoder : Pos -> JE.Value
posEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "x" JE.int 0 v.x)
        , (requiredFieldEncoder "y" JE.int 0 v.y)
        ]


type alias Piece =
    {
    }


type Piece_Id
    = Piece_Null -- 0
    | Piece_Gyoku -- 1
    | Piece_Hisha -- 2
    | Piece_Ryu -- 3
    | Piece_Kaku -- 4
    | Piece_Uma -- 5
    | Piece_Kin -- 6
    | Piece_Gin -- 7
    | Piece_NariGin -- 8
    | Piece_Kei -- 9
    | Piece_NariKei -- 10
    | Piece_Kyou -- 11
    | Piece_NariKyou -- 12
    | Piece_Fu -- 13
    | Piece_To -- 14


pieceDecoder : JD.Decoder Piece
pieceDecoder =
    JD.lazy <| \_ -> decode Piece


piece_IdDecoder : JD.Decoder Piece_Id
piece_IdDecoder =
    let
        lookup s =
            case s of
                "NULL" ->
                    Piece_Null

                "GYOKU" ->
                    Piece_Gyoku

                "HISHA" ->
                    Piece_Hisha

                "RYU" ->
                    Piece_Ryu

                "KAKU" ->
                    Piece_Kaku

                "UMA" ->
                    Piece_Uma

                "KIN" ->
                    Piece_Kin

                "GIN" ->
                    Piece_Gin

                "NARI_GIN" ->
                    Piece_NariGin

                "KEI" ->
                    Piece_Kei

                "NARI_KEI" ->
                    Piece_NariKei

                "KYOU" ->
                    Piece_Kyou

                "NARI_KYOU" ->
                    Piece_NariKyou

                "FU" ->
                    Piece_Fu

                "TO" ->
                    Piece_To

                _ ->
                    Piece_Null
    in
        JD.map lookup JD.string


piece_IdDefault : Piece_Id
piece_IdDefault = Piece_Null


pieceEncoder : Piece -> JE.Value
pieceEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


piece_IdEncoder : Piece_Id -> JE.Value
piece_IdEncoder v =
    let
        lookup s =
            case s of
                Piece_Null ->
                    "NULL"

                Piece_Gyoku ->
                    "GYOKU"

                Piece_Hisha ->
                    "HISHA"

                Piece_Ryu ->
                    "RYU"

                Piece_Kaku ->
                    "KAKU"

                Piece_Uma ->
                    "UMA"

                Piece_Kin ->
                    "KIN"

                Piece_Gin ->
                    "GIN"

                Piece_NariGin ->
                    "NARI_GIN"

                Piece_Kei ->
                    "KEI"

                Piece_NariKei ->
                    "NARI_KEI"

                Piece_Kyou ->
                    "KYOU"

                Piece_NariKyou ->
                    "NARI_KYOU"

                Piece_Fu ->
                    "FU"

                Piece_To ->
                    "TO"

    in
        JE.string <| lookup v


type alias FinishedStatus =
    {
    }


type FinishedStatus_Id
    = FinishedStatus_NotFinished -- 0
    | FinishedStatus_Suspend -- 1
    | FinishedStatus_Surrender -- 2
    | FinishedStatus_Draw -- 3
    | FinishedStatus_RepetitionDraw -- 4
    | FinishedStatus_Checkmate -- 5
    | FinishedStatus_OverTimeLimit -- 6
    | FinishedStatus_FoulLoss -- 7
    | FinishedStatus_FoulWin -- 8
    | FinishedStatus_NyugyokuWin -- 9


finishedStatusDecoder : JD.Decoder FinishedStatus
finishedStatusDecoder =
    JD.lazy <| \_ -> decode FinishedStatus


finishedStatus_IdDecoder : JD.Decoder FinishedStatus_Id
finishedStatus_IdDecoder =
    let
        lookup s =
            case s of
                "NOT_FINISHED" ->
                    FinishedStatus_NotFinished

                "SUSPEND" ->
                    FinishedStatus_Suspend

                "SURRENDER" ->
                    FinishedStatus_Surrender

                "DRAW" ->
                    FinishedStatus_Draw

                "REPETITION_DRAW" ->
                    FinishedStatus_RepetitionDraw

                "CHECKMATE" ->
                    FinishedStatus_Checkmate

                "OVER_TIME_LIMIT" ->
                    FinishedStatus_OverTimeLimit

                "FOUL_LOSS" ->
                    FinishedStatus_FoulLoss

                "FOUL_WIN" ->
                    FinishedStatus_FoulWin

                "NYUGYOKU_WIN" ->
                    FinishedStatus_NyugyokuWin

                _ ->
                    FinishedStatus_NotFinished
    in
        JD.map lookup JD.string


finishedStatus_IdDefault : FinishedStatus_Id
finishedStatus_IdDefault = FinishedStatus_NotFinished


finishedStatusEncoder : FinishedStatus -> JE.Value
finishedStatusEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


finishedStatus_IdEncoder : FinishedStatus_Id -> JE.Value
finishedStatus_IdEncoder v =
    let
        lookup s =
            case s of
                FinishedStatus_NotFinished ->
                    "NOT_FINISHED"

                FinishedStatus_Suspend ->
                    "SUSPEND"

                FinishedStatus_Surrender ->
                    "SURRENDER"

                FinishedStatus_Draw ->
                    "DRAW"

                FinishedStatus_RepetitionDraw ->
                    "REPETITION_DRAW"

                FinishedStatus_Checkmate ->
                    "CHECKMATE"

                FinishedStatus_OverTimeLimit ->
                    "OVER_TIME_LIMIT"

                FinishedStatus_FoulLoss ->
                    "FOUL_LOSS"

                FinishedStatus_FoulWin ->
                    "FOUL_WIN"

                FinishedStatus_NyugyokuWin ->
                    "NYUGYOKU_WIN"

    in
        JE.string <| lookup v


type alias Value =
    { name : String -- 1
    , value : String -- 2
    }


valueDecoder : JD.Decoder Value
valueDecoder =
    JD.lazy <| \_ -> decode Value
        |> required "name" JD.string ""
        |> required "value" JD.string ""


valueEncoder : Value -> JE.Value
valueEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "name" JE.string "" v.name)
        , (requiredFieldEncoder "value" JE.string "" v.value)
        ]


type alias Step =
    { seq : Int -- 1
    , position : String -- 2
    , src : Maybe Pos -- 3
    , dst : Maybe Pos -- 4
    , piece : Piece_Id -- 5
    , finishedStatus : FinishedStatus_Id -- 6
    , promoted : Bool -- 7
    , captured : Piece_Id -- 8
    , timestampSec : Int -- 9
    , thinkingSec : Int -- 10
    , notes : List String -- 11
    }


stepDecoder : JD.Decoder Step
stepDecoder =
    JD.lazy <| \_ -> decode Step
        |> required "seq" intDecoder 0
        |> required "position" JD.string ""
        |> optional "src" posDecoder
        |> optional "dst" posDecoder
        |> required "piece" piece_IdDecoder piece_IdDefault
        |> required "finishedStatus" finishedStatus_IdDecoder finishedStatus_IdDefault
        |> required "promoted" JD.bool False
        |> required "captured" piece_IdDecoder piece_IdDefault
        |> required "timestampSec" intDecoder 0
        |> required "thinkingSec" intDecoder 0
        |> repeated "notes" JD.string


stepEncoder : Step -> JE.Value
stepEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "seq" JE.int 0 v.seq)
        , (requiredFieldEncoder "position" JE.string "" v.position)
        , (optionalEncoder "src" posEncoder v.src)
        , (optionalEncoder "dst" posEncoder v.dst)
        , (requiredFieldEncoder "piece" piece_IdEncoder piece_IdDefault v.piece)
        , (requiredFieldEncoder "finishedStatus" finishedStatus_IdEncoder finishedStatus_IdDefault v.finishedStatus)
        , (requiredFieldEncoder "promoted" JE.bool False v.promoted)
        , (requiredFieldEncoder "captured" piece_IdEncoder piece_IdDefault v.captured)
        , (requiredFieldEncoder "timestampSec" JE.int 0 v.timestampSec)
        , (requiredFieldEncoder "thinkingSec" JE.int 0 v.thinkingSec)
        , (repeatedFieldEncoder "notes" JE.string v.notes)
        ]


type alias GetKifuResponse =
    { userId : String -- 1
    , kifuId : String -- 2
    , startTs : Int -- 3
    , endTs : Int -- 4
    , handicap : String -- 5
    , gameName : String -- 6
    , firstPlayers : List GetKifuResponse_Player -- 7
    , secondPlayers : List GetKifuResponse_Player -- 8
    , otherFields : List Value -- 9
    , sfen : String -- 10
    , createdTs : Int -- 11
    , steps : List Step -- 12
    , note : String -- 13
    }


getKifuResponseDecoder : JD.Decoder GetKifuResponse
getKifuResponseDecoder =
    JD.lazy <| \_ -> decode GetKifuResponse
        |> required "userId" JD.string ""
        |> required "kifuId" JD.string ""
        |> required "startTs" intDecoder 0
        |> required "endTs" intDecoder 0
        |> required "handicap" JD.string ""
        |> required "gameName" JD.string ""
        |> repeated "firstPlayers" getKifuResponse_PlayerDecoder
        |> repeated "secondPlayers" getKifuResponse_PlayerDecoder
        |> repeated "otherFields" valueDecoder
        |> required "sfen" JD.string ""
        |> required "createdTs" intDecoder 0
        |> repeated "steps" stepDecoder
        |> required "note" JD.string ""


getKifuResponseEncoder : GetKifuResponse -> JE.Value
getKifuResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "userId" JE.string "" v.userId)
        , (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        , (requiredFieldEncoder "startTs" numericStringEncoder 0 v.startTs)
        , (requiredFieldEncoder "endTs" numericStringEncoder 0 v.endTs)
        , (requiredFieldEncoder "handicap" JE.string "" v.handicap)
        , (requiredFieldEncoder "gameName" JE.string "" v.gameName)
        , (repeatedFieldEncoder "firstPlayers" getKifuResponse_PlayerEncoder v.firstPlayers)
        , (repeatedFieldEncoder "secondPlayers" getKifuResponse_PlayerEncoder v.secondPlayers)
        , (repeatedFieldEncoder "otherFields" valueEncoder v.otherFields)
        , (requiredFieldEncoder "sfen" JE.string "" v.sfen)
        , (requiredFieldEncoder "createdTs" numericStringEncoder 0 v.createdTs)
        , (repeatedFieldEncoder "steps" stepEncoder v.steps)
        , (requiredFieldEncoder "note" JE.string "" v.note)
        ]


type alias GetKifuResponse_Player =
    { name : String -- 1
    , note : String -- 2
    }


getKifuResponse_PlayerDecoder : JD.Decoder GetKifuResponse_Player
getKifuResponse_PlayerDecoder =
    JD.lazy <| \_ -> decode GetKifuResponse_Player
        |> required "name" JD.string ""
        |> required "note" JD.string ""


getKifuResponse_PlayerEncoder : GetKifuResponse_Player -> JE.Value
getKifuResponse_PlayerEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "name" JE.string "" v.name)
        , (requiredFieldEncoder "note" JE.string "" v.note)
        ]


type alias GetSamePositionsRequest =
    { userId : String -- 1
    , position : String -- 2
    , steps : Int -- 3
    }


getSamePositionsRequestDecoder : JD.Decoder GetSamePositionsRequest
getSamePositionsRequestDecoder =
    JD.lazy <| \_ -> decode GetSamePositionsRequest
        |> required "userId" JD.string ""
        |> required "position" JD.string ""
        |> required "steps" intDecoder 0


getSamePositionsRequestEncoder : GetSamePositionsRequest -> JE.Value
getSamePositionsRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "userId" JE.string "" v.userId)
        , (requiredFieldEncoder "position" JE.string "" v.position)
        , (requiredFieldEncoder "steps" JE.int 0 v.steps)
        ]


type alias GetSamePositionsResponse =
    { userId : String -- 1
    , position : String -- 2
    , kifus : List GetSamePositionsResponse_Kifu -- 3
    }


getSamePositionsResponseDecoder : JD.Decoder GetSamePositionsResponse
getSamePositionsResponseDecoder =
    JD.lazy <| \_ -> decode GetSamePositionsResponse
        |> required "userId" JD.string ""
        |> required "position" JD.string ""
        |> repeated "kifus" getSamePositionsResponse_KifuDecoder


getSamePositionsResponseEncoder : GetSamePositionsResponse -> JE.Value
getSamePositionsResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "userId" JE.string "" v.userId)
        , (requiredFieldEncoder "position" JE.string "" v.position)
        , (repeatedFieldEncoder "kifus" getSamePositionsResponse_KifuEncoder v.kifus)
        ]


type alias GetSamePositionsResponse_Kifu =
    { kifuId : String -- 1
    , seq : Int -- 2
    }


getSamePositionsResponse_KifuDecoder : JD.Decoder GetSamePositionsResponse_Kifu
getSamePositionsResponse_KifuDecoder =
    JD.lazy <| \_ -> decode GetSamePositionsResponse_Kifu
        |> required "kifuId" JD.string ""
        |> required "seq" intDecoder 0


getSamePositionsResponse_KifuEncoder : GetSamePositionsResponse_Kifu -> JE.Value
getSamePositionsResponse_KifuEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "kifuId" JE.string "" v.kifuId)
        , (requiredFieldEncoder "seq" JE.int 0 v.seq)
        ]


type alias KifuRequest =
    { kifuRequestSelect : KifuRequestSelect
    }


type KifuRequestSelect
    = KifuRequestSelectUnspecified
    | RequestRecentKifu RecentKifuRequest
    | RequestPostKifu PostKifuRequest
    | RequestDeleteKifu DeleteKifuRequest
    | RequestGetKifu GetKifuRequest
    | RequestGetSamePositions GetSamePositionsRequest


kifuRequestSelectDecoder : JD.Decoder KifuRequestSelect
kifuRequestSelectDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map RequestRecentKifu (JD.field "requestRecentKifu" recentKifuRequestDecoder)
        , JD.map RequestPostKifu (JD.field "requestPostKifu" postKifuRequestDecoder)
        , JD.map RequestDeleteKifu (JD.field "requestDeleteKifu" deleteKifuRequestDecoder)
        , JD.map RequestGetKifu (JD.field "requestGetKifu" getKifuRequestDecoder)
        , JD.map RequestGetSamePositions (JD.field "requestGetSamePositions" getSamePositionsRequestDecoder)
        , JD.succeed KifuRequestSelectUnspecified
        ]


kifuRequestSelectEncoder : KifuRequestSelect -> Maybe ( String, JE.Value )
kifuRequestSelectEncoder v =
    case v of
        KifuRequestSelectUnspecified ->
            Nothing
        RequestRecentKifu x ->
            Just ( "requestRecentKifu", recentKifuRequestEncoder x )
        RequestPostKifu x ->
            Just ( "requestPostKifu", postKifuRequestEncoder x )
        RequestDeleteKifu x ->
            Just ( "requestDeleteKifu", deleteKifuRequestEncoder x )
        RequestGetKifu x ->
            Just ( "requestGetKifu", getKifuRequestEncoder x )
        RequestGetSamePositions x ->
            Just ( "requestGetSamePositions", getSamePositionsRequestEncoder x )


kifuRequestDecoder : JD.Decoder KifuRequest
kifuRequestDecoder =
    JD.lazy <| \_ -> decode KifuRequest
        |> field kifuRequestSelectDecoder


kifuRequestEncoder : KifuRequest -> JE.Value
kifuRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [ (kifuRequestSelectEncoder v.kifuRequestSelect)
        ]


type alias KifuResponse =
    { kifuResponseSelect : KifuResponseSelect
    }


type KifuResponseSelect
    = KifuResponseSelectUnspecified
    | ResponseRecentKifu RecentKifuResponse
    | ResponsePostKifu PostKifuResponse
    | ResponseDeleteKifu DeleteKifuResponse
    | ResponseGetKifu GetKifuResponse
    | ResponseGetSamePositions GetSamePositionsResponse


kifuResponseSelectDecoder : JD.Decoder KifuResponseSelect
kifuResponseSelectDecoder =
    JD.lazy <| \_ -> JD.oneOf
        [ JD.map ResponseRecentKifu (JD.field "responseRecentKifu" recentKifuResponseDecoder)
        , JD.map ResponsePostKifu (JD.field "responsePostKifu" postKifuResponseDecoder)
        , JD.map ResponseDeleteKifu (JD.field "responseDeleteKifu" deleteKifuResponseDecoder)
        , JD.map ResponseGetKifu (JD.field "responseGetKifu" getKifuResponseDecoder)
        , JD.map ResponseGetSamePositions (JD.field "responseGetSamePositions" getSamePositionsResponseDecoder)
        , JD.succeed KifuResponseSelectUnspecified
        ]


kifuResponseSelectEncoder : KifuResponseSelect -> Maybe ( String, JE.Value )
kifuResponseSelectEncoder v =
    case v of
        KifuResponseSelectUnspecified ->
            Nothing
        ResponseRecentKifu x ->
            Just ( "responseRecentKifu", recentKifuResponseEncoder x )
        ResponsePostKifu x ->
            Just ( "responsePostKifu", postKifuResponseEncoder x )
        ResponseDeleteKifu x ->
            Just ( "responseDeleteKifu", deleteKifuResponseEncoder x )
        ResponseGetKifu x ->
            Just ( "responseGetKifu", getKifuResponseEncoder x )
        ResponseGetSamePositions x ->
            Just ( "responseGetSamePositions", getSamePositionsResponseEncoder x )


kifuResponseDecoder : JD.Decoder KifuResponse
kifuResponseDecoder =
    JD.lazy <| \_ -> decode KifuResponse
        |> field kifuResponseSelectDecoder


kifuResponseEncoder : KifuResponse -> JE.Value
kifuResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (kifuResponseSelectEncoder v.kifuResponseSelect)
        ]


type alias HelloRequest =
    {
    }


helloRequestDecoder : JD.Decoder HelloRequest
helloRequestDecoder =
    JD.lazy <| \_ -> decode HelloRequest


helloRequestEncoder : HelloRequest -> JE.Value
helloRequestEncoder v =
    JE.object <| List.filterMap identity <|
        [
        ]


type alias HelloResponse =
    { message : String -- 1
    , name : String -- 2
    }


helloResponseDecoder : JD.Decoder HelloResponse
helloResponseDecoder =
    JD.lazy <| \_ -> decode HelloResponse
        |> required "message" JD.string ""
        |> required "name" JD.string ""


helloResponseEncoder : HelloResponse -> JE.Value
helloResponseEncoder v =
    JE.object <| List.filterMap identity <|
        [ (requiredFieldEncoder "message" JE.string "" v.message)
        , (requiredFieldEncoder "name" JE.string "" v.name)
        ]
