network "dev" {
  subnet = "10.6.0.0/16"
}

module "consul_nomad" {
  source = "github.com/shipyard-run/blueprints//modules/consul-nomad"
}