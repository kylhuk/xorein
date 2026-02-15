let ConfigType = { application : Text, environment : Text, version : Text }

let Relay =
      let Node =
            { name : Text
            , region : Text
            , environment : Text
            , role : Text
            , listen : List Text
            , announce : List Text
            , tags : List Text
            }

      let Limits =
            { maxCircuits : Natural
            , maxCircuitDuration : Text
            , maxCircuitBandwidth : Text
            }

      let StoreForward =
            { enabled : Bool
            , storagePath : Text
            , maxMessages : Natural
            , maxBytes : Natural
            , ttl : Text
            }

      let Sfu =
            { enabled : Bool
            , maxRooms : Natural
            , maxParticipants : Natural
            }

      let Metrics =
            { enabled : Bool
            , listenAddr : Text
            }

      let Health = { interval : Text }

      let Config =
            { node : Node
            , limits : Limits
            , storeForward : StoreForward
            , sfu : Sfu
            , metrics : Metrics
            , health : Health
            }

      let Environment = { environment : Text, nodes : List Config }

      in  { Node = Node
          , Limits = Limits
          , StoreForward = StoreForward
          , Sfu = Sfu
          , Metrics = Metrics
          , Health = Health
          , Config = Config
          , Environment = Environment
          }

in  { ConfigType = ConfigType, Relay = Relay }
