module Pages.Home_ exposing (Model, Msg, page)

import Effect exposing (Effect)
import Gen.Params.Home_ exposing (Params)
import Html exposing (button, text)
import Html.Events as Events
import Page
import Request
import Shared
import Storage
import View exposing (View)


page : Shared.Model -> Request.With Params -> Page.With Model Msg
page shared req =
    Page.protected.advanced
        (\user ->
            { init = init
            , update = update
            , view = view user
            , subscriptions = subscriptions
            }
        )



-- INIT


type alias Model =
    {}


init : ( Model, Effect Msg )
init =
    ( {}, Effect.none )



-- UPDATE


type Msg
    = SignOut


update : Msg -> Model -> ( Model, Effect Msg )
update msg model =
    case msg of
        SignOut ->
            ( model, Effect.fromShared Shared.SignOut )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Storage.User -> Model -> View Msg
view user model =
    { title = "Home"
    , body =
        [ text "Hello, "
        , text user.token
        , text "!"
        , button [ Events.onClick SignOut ] [ text "Sign Out" ]
        ]
    }
