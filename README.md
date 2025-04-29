# argnewline

This tool reformats Go source files by converting single-line parameter lists into a multi-line format, improving readability and maintainability.

## Features

- **Function Declarations:** Converts single-line function parameter lists into multi-line lists with each parameter on its own line.
- **Function Calls:** Reformats call arguments when they are specified on a single line.
- **Interface Methods:** Adjusts interface method parameters from a single line to a multi-line layout.
- **Directory Scanning:** Processes individual files or recursively scans directories for `.go` files while skipping the `vendor` directory.
- **AST-Based:** Utilizes Go's Abstract Syntax Tree (AST) for precise modifications.
- **Automatic Formatting:** Applies Go's standard formatting using the `go/format` package after making modifications.

## Usage

1. **Build the Tool**

   Compile the program using:

   ```sh
   go build -o argnewline main.go
   ```

2. **Run the Tool**

   Execute the program with a file or directory as an argument:

   ```sh
   ./argnewline <path-or-file>
   ```

   - If a directory is specified, the tool will recursively process all `.go` files (excluding those in `vendor` directories).
   - If no path is provided, the tool displays a usage message and exits.

## How It Works

1. **Argument Parsing:**  
   The program reads command-line arguments to determine the target file or directory.

2. **File Collection:**  
   If a directory is provided, it walks through the directory tree, gathering `.go` files and skipping over `vendor` directories.

3. **AST Processing:**  
   For each target file:
   - The file is parsed into an AST using Go's parser.
   - The program inspects function declarations, function calls, and interface method definitions.
   - For nodes with parameters on a single line, it constructs a new multi-line format with proper indentation.
   
4. **Modification Application:**  
   Modifications are gathered and applied in reverse order (to account for offset changes) before the file is reformatted using Go's formatting standards.

5. **File Update:**  
   The updated content is written back to the original file.

## License

Refer to the [LICENSE](LICENSE) file for license details.
