let Defaults = ./default.dhall

let Types = ./types.dhall

let mkRelay =
      \(name : Text) ->
      \(region : Text) ->
      \(listen : List Text) ->
      \(announce : List Text) ->
      Defaults.relay.base
      // { node = Defaults.relay.base.node
            // { name = name, region = region, listen = listen, announce = announce }
          }

let mkBootstrap =
      \(name : Text) ->
      \(region : Text) ->
      \(listen : List Text) ->
      \(announce : List Text) ->
      \(contact : Text) ->
      Defaults.bootstrap.base
      // { node = Defaults.bootstrap.base.node
            // { name = name
               , region = region
               , listen = listen
               , announce = announce
               , contact = contact
               }
          }

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
    , bootstrapEnvironments =
        [ { environment = "dev"
            , nodes =
                [ mkBootstrap "bootstrap-dev-eu" "eu-west"
                    [ "/ip4/0.0.0.0/tcp/3101", "/ip4/0.0.0.0/udp/3101/quic-v1" ]
                    [ "/dns4/bootstrap-dev-eu.aether.test/tcp/3101" ]
                    "noc+dev-eu@aether.test"
                , mkBootstrap "bootstrap-dev-us" "us-east"
                    [ "/ip4/0.0.0.0/tcp/3201", "/ip4/0.0.0.0/udp/3201/quic-v1" ]
                    [ "/dns4/bootstrap-dev-us.aether.test/tcp/3201" ]
                    "noc+dev-us@aether.test"
                , mkBootstrap "bootstrap-dev-ap" "ap-south"
                    [ "/ip4/0.0.0.0/tcp/3301", "/ip4/0.0.0.0/udp/3301/quic-v1" ]
                    [ "/dns4/bootstrap-dev-ap.aether.test/tcp/3301" ]
                    "noc+dev-ap@aether.test"
                ]
          }
        , { environment = "staging"
            , nodes =
                [ mkBootstrap "bootstrap-staging-eu" "eu-central"
                    [ "/ip4/0.0.0.0/tcp/3401", "/ip4/0.0.0.0/udp/3401/quic-v1" ]
                    [ "/dns4/bootstrap-staging-eu.aether.test/tcp/3401" ]
                    "noc+staging-eu@aether.test"
                , mkBootstrap "bootstrap-staging-us" "us-west"
                    [ "/ip4/0.0.0.0/tcp/3501", "/ip4/0.0.0.0/udp/3501/quic-v1" ]
                    [ "/dns4/bootstrap-staging-us.aether.test/tcp/3501" ]
                    "noc+staging-us@aether.test"
                , mkBootstrap "bootstrap-staging-ap" "ap-south"
                    [ "/ip4/0.0.0.0/tcp/3601", "/ip4/0.0.0.0/udp/3601/quic-v1" ]
                    [ "/dns4/bootstrap-staging-ap.aether.test/tcp/3601" ]
                    "noc+staging-ap@aether.test"
                ]
          }
        , { environment = "production"
            , nodes =
                [ mkBootstrap "bootstrap-prod-eu" "eu-central"
                    [ "/ip4/0.0.0.0/tcp/3701", "/ip4/0.0.0.0/udp/3701/quic-v1" ]
                    [ "/dns4/bootstrap-prod-eu.aether.chat/tcp/3701" ]
                    "noc+prod-eu@aether.chat"
                , mkBootstrap "bootstrap-prod-us" "us-east"
                    [ "/ip4/0.0.0.0/tcp/3801", "/ip4/0.0.0.0/udp/3801/quic-v1" ]
                    [ "/dns4/bootstrap-prod-us.aether.chat/tcp/3801" ]
                    "noc+prod-us@aether.chat"
                , mkBootstrap "bootstrap-prod-ap" "ap-south"
                    [ "/ip4/0.0.0.0/tcp/3901", "/ip4/0.0.0.0/udp/3901/quic-v1" ]
                    [ "/dns4/bootstrap-prod-ap.aether.chat/tcp/3901" ]
                    "noc+prod-ap@aether.chat"
                ]
          }
        ] : List Types.Bootstrap.Environment
    }
