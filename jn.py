import sys
import os

def join(source_dir, dest_file, read_size):
    # Create a new destination file
    output_file = open(dest_file, 'wb')
     
    # Get a list of the file parts
    parts = os.listdir(source_dir)
     
    # Sort them by name (remember that the order num is part of the file name)
    parts.sort()
 
    # Go through each portion one by one
    for file in parts:
         
        # Assemble the full path to the file
        path = os.path.join(source_dir, file)
         
        # Open the part
        input_file = open(path, 'rb')
         
        while True:
            # Read all bytes of the part
            bytes = input_file.read(read_size)
             
            # Break out of loop if we are at end of file
            if not bytes:
                break
                 
            # Write the bytes to the output file
            output_file.write(bytes)
             
        # Close the input file
        input_file.close()
         
    # Close the output file
    output_file.close()


def usage():
    err_write(
    "Usage: [ python ] {} source_dir dest_file\n".format(
        sys.argv[0]))
    err_write(
    "joins in_filename into one file out_file\n".format(
        sys.argv[0]))

def main():

    if len(sys.argv) != 3:
        usage()
        sys.exit(1)

    try:
        join(sys.argv[1], sys.argv[2], 10000000) 
    except ValueError as ve:
        error_exit(str(ve))
    except IOError as ioe:
        error_exit(str(ioe))
    except Exception as e:
        error_exit(str(e))

if __name__ == '__main__':
    main()
