module Shared exposing
    ( Flags
    , Model
    , Msg(..)
    , init
    , subscriptions
    , update
    )

import Gen.Route
import Json.Decode as Json
import Request exposing (Request)
import Storage exposing (Storage)


type alias Flags =
    Json.Value


type alias Model =
    { storage : Storage }


type Msg
    = SignIn Storage.User
    | SignOut
    | StorageUpdated Storage


init : Request -> Flags -> ( Model, Cmd Msg )
init _ flags =
    ( { storage = Storage.fromJson flags }
    , Cmd.none
    )


update : Request -> Msg -> Model -> ( Model, Cmd Msg )
update req msg model =
    case msg of
        SignIn user ->
            ( model
            , Storage.signIn model.storage user.token
            )

        SignOut ->
            ( model, Storage.signOut model.storage )

        StorageUpdated storage ->
            case storage.user of
                Just _ ->
                    ( { model | storage = storage }, Request.pushRoute Gen.Route.Home_ req )

                Nothing ->
                    ( { model | storage = storage }, Cmd.none )


subscriptions : Request -> Model -> Sub Msg
subscriptions _ _ =
    Storage.onChange StorageUpdated
