port module Storage exposing (..)

import Json.Decode as Json
import Json.Encode as Encode


type alias User =
    { token : String }


type alias Storage =
    { user : Maybe User
    }


port save : Json.Value -> Cmd msg


port load : (Json.Value -> msg) -> Sub msg


toJson : Storage -> Json.Value
toJson storage =
    Encode.object
        [ ( "user"
          , case storage.user of
                Just user ->
                    Encode.object [ ( "token", Encode.string user.token ) ]

                Nothing ->
                    Encode.null
          )
        ]


fromJson : Json.Value -> Storage
fromJson value =
    value
        |> Json.decodeValue decoder
        |> Result.withDefault initial


decoder : Json.Decoder Storage
decoder =
    Json.map Storage
        (Json.field "user"
            (Json.maybe
                (Json.map User
                    (Json.field "token" Json.string)
                )
            )
        )


initial : Storage
initial =
    { user = Nothing
    }


onChange : (Storage -> msg) -> Sub msg
onChange fromStorage =
    load (\json -> fromJson json |> fromStorage)


signIn : Storage -> String -> Cmd msg
signIn storage token =
    { storage | user = Just { token = token } } |> toJson |> save


signOut : Storage -> Cmd msg
signOut storage =
    { storage | user = Nothing } |> toJson |> save
