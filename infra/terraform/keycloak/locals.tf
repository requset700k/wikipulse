locals {
  realm_roles = {
    student = {
      description = "Learner role for lab access and AI tutor usage."
    }
    instructor = {
      description = "Instructor role for learner creation and instructor mode."
    }
    admin = {
      description = "Operational administrator role."
    }
    observer = {
      description = "Read-only observer role for dashboards and logs."
    }
  }

  groups = {
    team-platform = {
      realm_roles = ["admin"]
    }
    team-security = {
      realm_roles = ["admin"]
    }
    team-observability = {
      realm_roles = ["admin", "observer"]
    }
    team-lab-data = {
      realm_roles = ["admin", "observer"]
    }
    team-ai = {
      realm_roles = ["admin", "observer"]
    }
    team-service = {
      realm_roles = ["admin", "observer"]
    }
    students-cohort-0 = {
      realm_roles = ["student"]
    }
    instructors = {
      realm_roles = ["instructor"]
    }
  }

  confidential_client_ids = [
    for client_id, client in var.oidc_clients : client_id
    if upper(client.access_type) == "CONFIDENTIAL"
  ]
}
