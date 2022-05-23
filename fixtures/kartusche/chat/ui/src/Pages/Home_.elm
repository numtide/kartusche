module Pages.Home_ exposing (Model, Msg, page)

import Gen.Params.Home_ exposing (Params)
import Html exposing (..)
import Http exposing (Error)
import Json.Decode exposing (list, string)
import Page
import Request
import Shared
import View exposing (View)


type Msg
    = GotChats (Result Http.Error (List String))


page : Shared.Model -> Request.With Params -> Page.With Model Msg
page shared req =
    Page.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }



-- INIT


type alias Model =
    { chats : List String
    }


init : ( Model, Cmd Msg )
init =
    ( { chats = [] }
    , Http.get
        { url = "/api/chats"
        , expect = Http.expectJson GotChats (list string)
        }
    )



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotChats res ->
            case res of
                Ok chats ->
                    ( { model | chats = chats }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> View Msg
view model =
    { title = "welcome to the chat"
    , body =
        [ ul
            []
            (List.map
                (\chat ->
                    li [] [ text chat ]
                )
                model.chats
            )
        ]
    }
