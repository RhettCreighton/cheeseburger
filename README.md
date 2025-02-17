# Cheeseburger CLI

Cheeseburger CLI is a simple command-line tool for hosting static onion sites via Tor. Designed for Linux, it streamlines the setup of a hidden service with minimal configuration. Future versions will expand functionality to host internally developed web applications.

## Overview

Cheeseburger CLI lets you:
- Host static websites as Tor hidden services.
- Generate customized (vanity) onion addresses.
- Operate in persistent mode using stored keys or in temporary mode.
- Launch secure hidden services with an embedded Tor binary.

## Installation and Requirements

- **Operating System:** Linux  
- **Tor Binary:** An embedded Tor binary (`bins/tor-linux-x86_64`) is used, so no separate installation is needed.

## Commands

### help
Displays usage instructions and available commands.

**Usage:**
```
cheeseburger help
```

### version
Shows the current version of the CLI.

**Usage:**
```
cheeseburger version
```

### vanity
Generates a vanity onion address by repeatedly generating key pairs until the address matches the specified prefix.

**Usage:**
```
cheeseburger vanity --prefix <desired_prefix> [--save] [--workers <num>]
```

**Options:**
- `--prefix` : Desired starting characters (in lowercase) for the onion address.
- `--save`   : Save the generated key details to disk (keys are stored under `data/vanity/default` or a custom directory).
- `--workers`: Number of parallel workers (defaults to the number of CPU cores).

When the `--save` flag is used, the following files are created or updated:
- `hs_ed25519_secret_key` – Contains the secret key.
- `hs_ed25519_public_key` – Contains the public key.
- `hostname` – Contains the onion address (with a `.onion` suffix).

### serve
Launches the Tor hidden service along with a static file server.

**Usage:**
```
cheeseburger serve <static_directory> [--vanity-name <name>]
```

**Parameters:**
- `<static_directory>`: The directory containing static site files.
- `--vanity-name`: (Optional) Load a custom vanity key set from `data/vanity/<vanity-name>/vanity.json`. If omitted, the default key set is used.

Under the hood, the `serve` command:
- Determines the hidden service directory.
- Validates and loads existing key files (if available).
- Configures file permissions and creates a `torrc` configuration for Tor.
- Starts a static HTTP server on port 8080 to serve your content.
- Executes the Tor process to establish the hidden service and display the generated onion address.

## Usage Examples

- **Display Help:**
  ```
  cheeseburger help
  ```
- **Show Version:**
  ```
  cheeseburger version
  ```
- **Generate a Vanity Onion Address (without saving):**
  ```
  cheeseburger vanity --prefix test
  ```
- **Generate and Save a Vanity Onion Address:**
  ```
  cheeseburger vanity --prefix test --save
  ```
- **Serve Static Files (default vanity key):**
  ```
  cheeseburger serve static-site/
  ```
- **Serve Static Files (custom vanity key):**
  ```
  cheeseburger serve static-site/ --vanity-name custom
  ```

## Debugging

- Review log messages for details on key generation, file validations, permissions, and Tor process status.
- Ensure that the hidden service directory and key files have correct permissions.

## Future Enhancements

- Support for hosting internally developed web applications.
- Additional configuration options and command extensions.

## License

This project is licensed under the MIT License.

```
MIT License

Copyright (c) 2025 Cheeseburger CLI

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
