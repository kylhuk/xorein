let Types = ./types.dhall

let baseNode : Types.Relay.Node =
      { name = "relay-template"
      , region = "global"
      , environment = "dev"
      , role = "relay"
      , listen =
          [ "/ip4/0.0.0.0/tcp/4001"
          , "/ip4/0.0.0.0/udp/4001/quic-v1"
          , "/ip4/0.0.0.0/tcp/4002/ws"
          ]
      , announce = [ "/dns4/relay-template.aether.test/tcp/4001" ]
      , tags = [ "relay", "prototype" ]
      }

let baseLimits : Types.Relay.Limits =
      { maxCircuits = 512
      , maxCircuitDuration = "2m"
      , maxCircuitBandwidth = "1Mbps"
      }

let baseStoreForward : Types.Relay.StoreForward =
      { enabled = True
      , storagePath = "/var/lib/aether/store"
      , maxMessages = 8192
      , maxBytes = 536870912
      , ttl = "30d"
      }

let baseSfu : Types.Relay.Sfu =
      { enabled = True
      , maxRooms = 50
      , maxParticipants = 100
      }

let baseMetrics : Types.Relay.Metrics =
      { enabled = True
      , listenAddr = "0.0.0.0:9090"
      }

let baseHealth : Types.Relay.Health = { interval = "30s" }

let baseRelayConfig : Types.Relay.Config =
      { node = baseNode
      , limits = baseLimits
      , storeForward = baseStoreForward
      , sfu = baseSfu
      , metrics = baseMetrics
      , health = baseHealth
      }

in  { ConfigType = Types.ConfigType
    , default = { application = "aether", environment = "dev", version = "0.1" }
    , relay = { base = baseRelayConfig }
    }
