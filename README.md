# wsfetch

A command-line tool for fetching and displaying Wealthsimple account information and transactions.

## Overview

`wsfetch` is a Go-based CLI tool that interacts with the Wealthsimple API to retrieve your account information and transaction history. It uses GraphQL queries to fetch data and provides a convenient way to view your financial information directly from the terminal.

## Features

- Authentication with Wealthsimple using username/password
- Session management (saves authentication tokens for future use)
- Retrieval of all account information
- Fetching transaction history with detailed descriptions
- Support for various transaction types:
  - Deposits and withdrawals
  - Transfers between accounts
  - Stock purchases and sales
  - Dividends
  - Interest payments
  - Currency conversions
  - Bill payments
  - Peer-to-peer payments
  - And more

## Installation

### Prerequisites

- Go 1.16 or higher

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/vpineda1996/wsfetch.git
   cd wsfetch
   ```

2. Build the binary:
   ```
   make build
   ```

3. (Optional) Install the binary to your PATH:
   ```
   make install
   ```

## Usage

### Basic usage

```
wsfetch fetch
```

This will:
1. Prompt for your Wealthsimple username and password if no saved session exists
2. Retrieve all your account information
3. Fetch transactions from the last 30 days
4. Display the information in a formatted JSON output

### Authentication

The first time you run `wsfetch`, it will prompt you for your Wealthsimple credentials:

```
Enter your username:
Enter your password:
```

After successful authentication, the session is saved to a local file (`session.json`) for future use, so you don't need to enter your credentials each time.

## Example Output

When running `wsfetch fetch`, you'll see output similar to:

```
fetch called
Enter your username:
your.email@example.com
Enter your password:
********
Account: TFSA, ID: account-123456
Account: PERSONAL, ID: account-789012
[
  {
    "Date": "2025-04-01T10:30:45Z",
    "Merchant": "ACME Corp",
    "Category": "DIVIDEND",
    "Account": "account-123456",
    "OriginalStatement": "",
    "Amount": "25.50",
    "Description": "Dividend: AAPL"
  },
  {
    "Date": "2025-03-28T14:22:10Z",
    "Merchant": "unknown",
    "Category": "DIY_BUY",
    "Account": "account-123456",
    "OriginalStatement": "",
    "Amount": "-500.00",
    "Description": "DIY BUY: buy 5 x AAPL @ 100.00"
  }
]
```

## Development

### Project Structure

- `cmd/`: Command-line interface definitions
  - `fetch.go`: Implementation of the fetch command
  - `root.go`: Root command definition
- `pkg/`: Package code
  - `client/`: Wealthsimple API client implementation
  - `auth/`: Authentication handling
  - `base/`: Base HTTP client functionality
  - `endpoints/`: API endpoint definitions

### Building

```
make build
```

### Testing

```
make test
```

## License

This project is licensed under the terms found in the [LICENSE](LICENSE) file.

## Disclaimer

This is an unofficial tool and is not affiliated with, maintained, authorized, endorsed, or sponsored by Wealthsimple.