let Types = ./types.dhall

in  { ConfigType = Types.ConfigType, default = { application = "aether", environment = "dev", version = "0.1" } }
