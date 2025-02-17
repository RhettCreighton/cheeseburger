# Cheeseburger

Cheeseburger is a Tor v3 Onion Hosting app that allows you to quickly generate vanity outputs and serve a static site.

## Quickstart

To get started with Cheeseburger, simply copy and paste these commands in your terminal. Modify the examples as needed for your own setup.

1. Generate vanity outputs:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger vanity --prefix test --save
   2025/02/18 01:20:10 Total Attempts: 2000000
   2025/02/18 01:20:13 Total Attempts: 3000000
   ```
   This command uses the `vanity` subcommand with a prefix option (in this example, "test") and saves the generated output.

2. Serve your static site:
   ```
   bob@ltp:~/projects/cheeseburger$ ./cheeseburger serve ./static-site/
   ```
   This command launches a simple server to serve the contents of your `./static-site/` directory.

## Additional Information

- Ensure you are in the project root directory: `/home/bob/projects/cheeseburger`.
- For further options and detailed usage, try running:
  ```
  ./cheeseburger --help
  ```
- Adjust the commands and options as needed to fit your project requirements.

## License

This project is licensed under the terms specified in the LICENSE file.
