# Data Architecture — System vs User Data

## Principle
All platform-provided ("system") data is clearly separated from user-created data via flags and conventions. System data ships with the product and is never modified by users. Users create their own data, optionally cloning from system templates.

## Separation by Table

### Frameworks & Controls (Sprint 2)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `frameworks` | `is_custom = false` | 5 built-in (SOC 2, ISO 27001, PCI DSS, GDPR, CCPA) | Custom frameworks (`is_custom = true`) |
| `framework_versions` | Inherits from parent framework | Pre-built versions with requirements | User-created versions |
| `requirements` | Inherits from framework | Pre-built requirements per framework version | User additions to custom frameworks |
| `controls` | `is_custom = false` | 275 pre-mapped controls | Custom controls (`is_custom = true`) |
| `control_mappings` | — | Pre-built cross-framework mappings | User-created mappings |

### Policies (Sprint 5)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `policies` | `is_template = true` | 13 policy templates (per framework) | User policies (`is_template = false`) |
| `policy_versions` | Inherits from parent policy | Template initial versions | User-authored versions |

### Risk Register (Sprint 6)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `risks` | `is_template = true` | 230 risk templates across 13 categories | User risks (`is_template = false`) |
| `risk_assessments` | — | None (always user data) | All assessments |
| `risk_treatments` | — | None (always user data) | All treatment plans |
| `risk_controls` | — | None (always user data) | All risk-control links |

### Audit Hub (Sprint 7)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `audit_request_templates` | Entire table is system data | 80 PBC templates (SOC 2, PCI, ISO, GDPR) | None — users don't modify templates |
| `audits` | — | None (always user data) | All audit engagements |
| `audit_requests` | `source_template_id` (nullable) | None | All requests (may reference template origin) |
| `audit_findings` | — | None (always user data) | All findings |

### Evidence (Sprint 3)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `evidence_artifacts` | — | None (always user data) | All uploaded evidence |

### Monitoring (Sprint 4)
| Table | Flag | System Data | User Data |
|-------|------|-------------|-----------|
| `tests` | `source_system` | Pre-built test definitions | User-created tests |
| `alert_rules` | — | Pre-built default rules | User-created rules |

## API Behavior

- **List endpoints** default to showing **user data only** (e.g., `GET /risks` returns `is_template=false`)
- Templates are accessed via explicit parameter (e.g., `?is_template=true`) or dedicated endpoints
- Create-from-template endpoints clone system data into user data with proper attribution (`template_source_id`, `source_template_id`)
- System data is **read-only** — users cannot edit or delete built-in frameworks, controls, or templates

## Seed Data Files

System data is loaded via numbered migration files:
- `001-026`: Schema + Sprint 1-4 seed data
- `033_sprint5_seed_templates.sql`: Policy templates
- `034_sprint5_seed_demo.sql`: Demo policies (for testing only)
- `042_sprint6_seed_templates.sql`: Risk templates (230)
- `043_sprint6_seed_demo.sql`: Demo risks (for testing only)
- `051_sprint7_seed_templates.sql`: PBC audit request templates (80)
- `052_sprint7_seed_demo.sql`: Demo audit engagement (for testing only)

### Convention
- `*_seed_templates.sql` = **System data** (ships with product)
- `*_seed_demo.sql` = **Demo data** (for development/testing, excluded in production)

## Future Considerations

1. **Versioned system data**: When frameworks update (e.g., PCI DSS v4.1), add new `framework_version` rows without modifying existing user data
2. **Tenant-scoped templates**: Allow orgs to create their own template library (templates visible only within their org)
3. **Template marketplace**: Share templates across organizations
4. **Migration strategy**: System data updates via new migration files, never modifying existing rows
