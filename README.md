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
   2025/02/18 01:20:15 Found matching address: testxyz...onion
   2025/02/18 01:20:15 Keys saved to: data/keys/testxyz
   ```
   This command uses the `vanity` subcommand with a prefix option (e.g., "test") and saves the generated output to the data/keys directory.

2. Serve your static site:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger serve ./static-site/
   2025/02/18 01:21:00 Starting Tor service...
   2025/02/18 01:21:05 Onion service running at: xyz123...onion
   2025/02/18 01:21:05 Serving static files from: ./static-site/
   ```
   This command launches a Tor hidden service to serve the contents of your `./static-site/` directory.

## Database Management Commands

Cheeseburger includes several commands for managing the application database:

1. Initialize a new database:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger mvc init
   Database initialized successfully
   ```
   Creates a new empty database for storing application data.

2. Create database backup:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger mvc backup
   Database backed up successfully to data/backups/backup_1708365245.db
   ```
   Generates a timestamped backup file in the data/backups directory.

3. Restore from backup:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger mvc restore data/backups/backup_1708365245.db
   Existing database found. Do you want to replace it? [y/N] y
   Database restored successfully
   ```
   Restores the database from a previous backup file.

4. Clean database:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger mvc clean
   Are you sure you want to clean the database? This cannot be undone. [y/N] y
   Database cleaned successfully
   ```
   Removes the existing database. Use with caution as this operation cannot be undone.

## MVC Application Commands

Run the dynamic blog application as a Tor hidden service:

```
bob@ltp:~/projects/cheeseburger$ ./cheeseburger mvc serve --vanity-name myblog
2025/02/18 01:22:00 Starting Tor service...
2025/02/18 01:22:05 Loading vanity keys from: data/keys/myblog
2025/02/18 01:22:05 Blog service running at: myblog...onion
2025/02/18 01:22:05 Database connected successfully
```

The `--vanity-name` option allows you to use a previously generated vanity address for your blog service.

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
  or
  ```
  ./cheeseburger mvc help
  ```
- Adjust the commands and options as needed to suit your project requirements.

## License

This project is licensed under the terms specified in the LICENSE file.
