job "dp-import-api" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  // Make sure that this API is only ran on the publishing nodes
  constraint {
    attribute = "${node.class}"
    value     = "publishing"
  }

  group "publishing" {
    count = {{PUBLISHING_TASK_COUNT}}

    task "dp-import-api" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/ons-dp-deployments/dp-import-api/latest.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

         args = [
                  "${NOMAD_TASK_DIR}/dp-import-api",
                ]
      }

      service {
        name = "dp-import-api"
        tags = ["publishing"]
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-import-api-publishing"]
      }
    }
  }
}