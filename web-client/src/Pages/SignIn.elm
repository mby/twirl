module Pages.SignIn exposing (Model, Msg, page)

import Effect exposing (Effect)
import Gen.Params.SignIn exposing (Params)
import Html exposing (br, button, input, text)
import Html.Events as Events
import Page
import Request
import Shared
import View exposing (View)


page : Shared.Model -> Request.With Params -> Page.With Model Msg
page shared req =
    Page.advanced
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }



-- INIT


type alias Model =
    { username : String
    , password : String
    }


init : ( Model, Effect Msg )
init =
    ( { username = ""
      , password = ""
      }
    , Effect.none
    )



-- UPDATE


type Msg
    = ClickedSignIn
    | UpdateUsername String
    | UpdatePassword String


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        ClickedSignIn ->
            ( model
            , Effect.fromShared (Shared.SignIn { token = "token" })
            )

        UpdateUsername username ->
            ( { model | username = username }, Effect.none )

        UpdatePassword password ->
            ( { model | password = password }, Effect.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "Sign In"
    , body =
        [ input [ Events.onInput UpdateUsername ] [ text model.username ]
        , br [] []
        , input [ Events.onInput UpdatePassword ] [ text model.password ]
        , br [] []
        , button [ Events.onClick ClickedSignIn ] [ text "Sign In" ]
        ]
    }
