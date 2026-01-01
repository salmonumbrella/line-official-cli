# LINE Ads API Research

## Executive Summary

The LINE Ads API (also called LINE Ads Management API) is a separate API from the LINE Messaging API, with different authentication mechanisms, base URLs, and access requirements. It is **not** included in the official `line/line-openapi` repository and requires corporate partnership approval to access.

## Authentication

### Method: JWS (JSON Web Signature) with HMAC-SHA-256

The LINE Ads API uses a **completely different authentication mechanism** from the Messaging API. Instead of channel access tokens, it uses JWS-based request signing.

#### Credential Acquisition
- Access Key and Secret Key are obtained from the Group settings page in LINE Ads Manager
- Requires an invitation email to join an API-enabled Group
- Credentials are tied to corporate partner relationships

#### Request Signing Process

1. **JOSE Header** (JSON):
   ```json
   {
     "alg": "HS256",
     "kid": "<access-key>",
     "typ": "text/plain"
   }
   ```

2. **Payload** (concatenated string):
   - SHA-256 digest of request body (hex encoded)
   - Content-Type header value
   - Date in `YYYYMMDD` format
   - Canonical URI path

3. **Signature Calculation**:
   ```
   InputValue = Base64(Header) + "." + Base64(Payload)
   Signature = Base64(HMAC-SHA-256(secretKey, InputValue))
   FinalToken = InputValue + "." + Signature
   ```

4. **Required HTTP Headers**:
   ```
   Content-Type: application/json
   Date: Wed, 22 Dec 2021 00:00:00 GMT
   Authorization: Bearer <FinalToken>
   ```

### Key Difference from Messaging API

| Aspect | Messaging API | Ads API |
|--------|--------------|---------|
| Auth Method | Channel Access Token (Bearer) | JWS Signature (HMAC-SHA-256) |
| Token Source | LINE Developers Console | LINE Ads Manager |
| Token Type | Long-lived or stateless JWT | Request-specific signature |
| Base URL | `api.line.me` | `ads.line.me` |

## Key Endpoints

**Base URL**: `https://ads.line.me/api`
**Current Version**: v3.11.3.1

### Groups Management
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/groups/{groupId}/children` | List child groups |
| POST | `/v3/groups/{groupId}/children` | Create child group |
| POST | `/v3/groups/{groupId}` | Update group |

### Link Requests (Partner Linking)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/groups/{groupId}/link-request` | Read link requests |
| POST | `/v3/groups/{groupId}/link-request/adaccount` | Create link request |
| POST | `/v3/groups/{groupId}/link-request/{id}/{actionType}` | Update link request |

### Ad Accounts
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/groups/{groupId}/adaccounts` | List authorized ad accounts |

### Campaigns
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/campaigns` | List campaigns |

### Ad Groups
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/adgroups` | List ad groups |

### Ads
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/ads` | List ads |

### Media Assets
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/media` | List media assets |

### Reports (Async CSV Download)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/pfreports` | List performance reports |
| POST | `/v3/adaccounts/{adaccountId}/pfreports` | Create performance report |
| DELETE | `/v3/adaccounts/{adaccountId}/pfreports/{id}` | Delete report |
| GET | `/v3/adaccounts/{adaccountId}/pfreports/{id}/download` | Download CSV report |

### Online Reports (Sync JSON Response)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/adaccounts/{adaccountId}/reports/online/{reportLevel}` | Get report as JSON |

### Custom Audiences
| Method | Path | Description |
|--------|------|-------------|
| Various | `/v3/adaccounts/{adaccountId}/customaudiences/*` | Manage custom audiences |

### Custom Conversions
| Method | Path | Description |
|--------|------|-------------|
| Various | `/v3/adaccounts/{adaccountId}/customconversions/*` | Manage custom conversions |

### Product Catalog
| Method | Path | Description |
|--------|------|-------------|
| Various | `/v3/adaccounts/{adaccountId}/productsets/*` | Manage product sets |
| Various | `/v3/adaccounts/{adaccountId}/products/*` | Manage products |

### Reference Codes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/v3/codes/ssps` | SSP codes |
| GET | `/v3/codes/placements` | Placement codes |
| GET | `/v3/codes/ages` | Age demographic codes |
| GET | `/v3/codes/genders` | Gender codes |
| GET | `/v3/codes/os` | Operating system codes |
| GET | `/v3/codes/call-to-actions` | CTA codes |

### Simulations
| Method | Path | Description |
|--------|------|-------------|
| Various | `/v3/.../simulations/*` | Ad group reach/frequency forecasting |

## Rate Limits

### API Request Rate
- **General Rate Limit**: 2 requests per second per API user

### Online Reports API
- **Concurrent Request Limit**: ~30 concurrent requests
- **Recommended**: 20 concurrent requests (for 3-second response times)

### HTTP Status Codes
| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad request |
| 401 | Invalid authorization token |
| 403 | API authorization missing |
| 429 | Rate limit exceeded |
| 500 | Server error |

## OpenAPI Availability

### Official OpenAPI Spec: NOT AVAILABLE

The LINE Ads API does **not** have a publicly available OpenAPI specification:

1. **line/line-openapi repository**: Only covers `api.line.me`, `api-data.line.me`, and `manager.line.biz` endpoints. The Ads API at `ads.line.me` is explicitly excluded.

2. **ads.line.me documentation**: Provides HTML documentation only, no downloadable OpenAPI/Swagger specs.

3. **Partner-only access**: OpenAPI specs may be available to approved partners, but this would require corporate partnership application.

### Implications for CLI Integration

Without an OpenAPI spec, client code must be:
- Written manually based on documentation
- Generated from documentation using custom tooling
- Obtained through partnership with LINE (if available to partners)

## Access Requirements

### Who Can Use the API

The LINE Ads API is **restricted to corporate users only**:

1. Must be an approved corporate partner
2. Must apply through LINE Ads Manager
3. Must be classified as one of:
   - **Certified Ad Tech General Partner**
   - **Reporting Partner**
   - **Data Provider Partner**

### Application Process

1. Contact LINE representative or submit inquiry through LINE for Business
2. Complete corporate partnership application
3. Receive invitation email to API-enabled Group
4. Access credentials from LINE Ads Manager Group settings

## Integration Recommendations

### Should LINE Ads API Be Included in the CLI?

**Recommendation: DEFER or OPTIONAL MODULE**

#### Reasons to Defer

1. **Different Authentication Model**: The JWS signature-based auth is completely different from the Messaging API's token-based auth. This adds significant complexity.

2. **Restricted Access**: Only corporate partners with LINE Ads accounts can use this API. Most CLI users would not have access.

3. **No OpenAPI Spec**: Without an OpenAPI spec, code generation is not possible. Manual implementation would be required and harder to maintain.

4. **Different User Base**: Messaging API users (developers building chatbots) vs Ads API users (marketing/advertising teams) have different needs.

#### If Integration is Desired

1. **Separate Authentication Flow**:
   ```
   line auth login --ads   # Different auth flow for Ads API
   ```

2. **Modular Design**:
   ```go
   // internal/ads/client.go - Separate client with JWS signing
   type AdsClient struct {
       accessKey  string
       secretKey  string
       httpClient *http.Client
   }

   func (c *AdsClient) sign(req *http.Request) error {
       // Implement JWS signature calculation
   }
   ```

3. **Separate Command Group**:
   ```
   line ads campaigns list --account <id>
   line ads reports create --account <id> --type performance
   line ads reports download <report-id>
   ```

4. **Configuration**:
   ```yaml
   # ~/.line-cli/config.yaml
   messaging:
     channel_id: "123456"
     channel_secret: "abc..."

   ads:
     access_key: "ak_..."
     secret_key: "sk_..."
   ```

### Suggested Approach

1. **Phase 1**: Complete Messaging API integration first
2. **Phase 2**: Add Ads API as optional plugin/module if there's user demand
3. **Phase 3**: Consider partnership with LINE for OpenAPI spec access

## References

- [LINE Ads API Overview (LINE Developers)](https://developers.line.biz/en/docs/line-ads-api/)
- [LINE Ads API About Page](https://developers.line.biz/en/docs/line-ads-api/about/)
- [LINE Ads Management API Specification (Partner Docs)](https://ads.line.me/public-docs/pages/v3/3.11.0/certificated-ad-tech-general-partner/)
- [LINE OpenAPI Repository (GitHub)](https://github.com/line/line-openapi)
