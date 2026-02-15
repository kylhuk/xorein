let Defaults = ./default.dhall

let Types = ./types.dhall

let mkRelay =
      \name : Text ->
      \region : Text ->
      \listen : List Text ->
      \announce : List Text ->
      Defaults.relay.base //\ { node = Defaults.relay.base.node // { name = name, region = region, listen = listen, announce = announce } }

in  { config = Defaults.default
    , relayEnvironments =
        [ { environment = "dev"
          , nodes =
              [ mkRelay "relay-dev-eu" "eu-west"
                  [ "/ip4/0.0.0.0/tcp/4101", "/ip4/0.0.0.0/udp/4101/quic-v1", "/ip4/0.0.0.0/tcp/4102/ws" ]
                  [ "/dns4/relay-dev-eu.aether.test/tcp/4101" ]
              , mkRelay "relay-dev-us" "us-east"
                  [ "/ip4/0.0.0.0/tcp/4201", "/ip4/0.0.0.0/udp/4201/quic-v1", "/ip4/0.0.0.0/tcp/4202/ws" ]
                  [ "/dns4/relay-dev-us.aether.test/tcp/4201" ]
              ]
          }
        , { environment = "staging"
          , nodes =
              [ mkRelay "relay-staging-eu" "eu-central"
                  [ "/ip4/0.0.0.0/tcp/4301", "/ip4/0.0.0.0/udp/4301/quic-v1", "/ip4/0.0.0.0/tcp/4302/ws" ]
                  [ "/dns4/relay-staging-eu.aether.test/tcp/4301" ]
              , mkRelay "relay-staging-us" "us-west"
                  [ "/ip4/0.0.0.0/tcp/4401", "/ip4/0.0.0.0/udp/4401/quic-v1", "/ip4/0.0.0.0/tcp/4402/ws" ]
                  [ "/dns4/relay-staging-us.aether.test/tcp/4401" ]
              ]
          }
        , { environment = "production"
          , nodes =
              [ mkRelay "relay-prod-eu" "eu-central"
                  [ "/ip4/0.0.0.0/tcp/4501", "/ip4/0.0.0.0/udp/4501/quic-v1", "/ip4/0.0.0.0/tcp/4502/ws" ]
                  [ "/dns4/relay-prod-eu.aether.chat/tcp/4501" ]
              , mkRelay "relay-prod-us" "us-east"
                  [ "/ip4/0.0.0.0/tcp/4601", "/ip4/0.0.0.0/udp/4601/quic-v1", "/ip4/0.0.0.0/tcp/4602/ws" ]
                  [ "/dns4/relay-prod-us.aether.chat/tcp/4601" ]
              , mkRelay "relay-prod-ap" "ap-south"
                  [ "/ip4/0.0.0.0/tcp/4701", "/ip4/0.0.0.0/udp/4701/quic-v1", "/ip4/0.0.0.0/tcp/4702/ws" ]
                  [ "/dns4/relay-prod-ap.aether.chat/tcp/4701" ]
              ]
          }
        ] : List Types.Relay.Environment
    }
