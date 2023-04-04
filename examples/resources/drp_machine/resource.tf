resource "drp_machine" "one_random_node" {
  pool                = k8s_pool
  allocate_workflow   = universal_k8s_build
  deallocate_workflow = universal_k8s_decom
  timeout             = "120m"
  add_profiles        = ["admin_access_keys", "k8s_node_network_settings"]
  filters             = ["Address=Ne()"]
}