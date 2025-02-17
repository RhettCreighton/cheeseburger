# Cheeseburger

Cheeseburger allows you to create your own Tor v3 Onion Hidden Service.

## Features

- Built-in Vanity Domain Name Generator
- Static Site Hosting
- Dynamic App Hosting
- Quantum Secure Replicate State Machine
- Privacy-by-Default P2P Messaging System (Quantum Secure)

## Quickstart

To get started with Cheeseburger, copy and paste the following commands into your terminal. Modify the examples as needed for your own setup.

1. Generate vanity outputs:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger vanity --prefix test --save
   2025/02/18 01:20:10 Total Attempts: 2000000
   2025/02/18 01:20:13 Total Attempts: 3000000
   ```
   This command uses the `vanity` subcommand with a prefix option (e.g., "test") and saves the generated output.

2. Serve your static site:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger serve ./static-site/
   ```
   This command launches a simple server to serve the contents of your `./static-site/` directory.

## Dependencies

Cheeseburger requires the following Linux dependency:

- libevent-2.1

### Installation Examples

- Debian/Ubuntu:
  ```
  sudo apt-get install libevent-2.1
  ```
- Fedora:
  ```
  sudo dnf install libevent
  ```
- Arch Linux:
  ```
  sudo pacman -S libevent
  ```

For other distributions, refer to your package manager or your distribution's repository for the appropriate package name.

## Additional Information

- Ensure you are in the project root directory: `/home/bob/projects/cheeseburger`.
- For further options and detailed usage, run:
  ```
  ./cheeseburger --help
  ```
- Adjust the commands and options as needed to suit your project requirements.

## License

This project is licensed under the terms specified in the LICENSE file.
