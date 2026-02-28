# Configuration

The server needs a single credential: a Google PageSpeed Insights API key. It can be provided in three ways.

---

## Credential

### `GOOGLE_PSI_API_KEY`

Your PageSpeed Insights API key from the Google Cloud Console.

| Method | Example |
|--------|---------|
| Environment variable | `GOOGLE_PSI_API_KEY=AIzaXXX` |
| CLI argument | `--api-key AIzaXXX` |
| `.env` file | `GOOGLE_PSI_API_KEY=AIzaXXX` |

---

## Resolution Order

The server resolves the API key using this priority (highest to lowest):

1. **CLI argument** (`--api-key`)
2. **Environment variable** (`GOOGLE_PSI_API_KEY` in process env or tool config `env` block)
3. **`.env` file** in the working directory

The first source that provides a non-empty value is used. If none are present, the server exits with an error at startup.

---

## No Additional Configuration

Unlike OAuth2-based servers, there is no token refresh, no service account JSON, and no billing account to configure. The PSI API key is the only credential.

- **Free quota**: 25,000 requests per day per Google Cloud project -- far more than any interactive use requires.
- **Billing**: Not required. Enable the API on a free-tier project and it works.
- **Restricting the key**: In Google Cloud Console → Credentials, restrict the key to the PageSpeed Insights API only to limit exposure.

