# Perfect rectangle packer
This is a program to find solutions for instances of the perfect rectangle packing problem. 

## Build
No additional dependencies are required. simply run ```go build``` in the main folder.

## Usage
A minimal example would be:
```
./tilingsolver -solver_id 1 -input_file testinputs.csv -output_dir ./output_log_directory
```

## Input format
The program reads a csv file with the following fields as input:
* ```job_id``` and ```puzzle_id``` should be integers and are only used as identifiers to connect status and solutions to a specific puzzle and job.
* ```num_tiles```, ```board_with``` and ```board_height```, how many tiles in the puzzle, and the board dimensions all as integers.
* ```tiles``` is a json encoded array of objects with an X and Y dimension for each tile. For best performance the tiles should roughly be sorted from large to small. It expects the X field to contain the largest side of a tile.
* ```start``` and ```end``` are used to specify where a job should start or end. If unused it should be an empty string, otherwise a json encoded array of up to ```num_tiles``` elements, in the order they should be placed in, with ```Idx``` referencing a tile index as ordered in ```tiles```, and ```rot``` a boolean, true if the tile was placed 90 degrees rotated.

```
"job_id","puzzle_id","num_tiles","board_width","board_height","tiles","start","end"
"1","1","6","9","9","[{""X"":6,""Y"":4},{""X"":6,""Y"":2},{""X"":5,""Y"":3},{""X"":5,""Y"":2},{""X"":4,""Y"":3},{""X"":4,""Y"":2}]","",""
"2","2","6","10","6","[{""X"":6,""Y"":2},{""X"":5,""Y"":4},{""X"":5,""Y"":1},{""X"":4,""Y"":3},{""X"":4,""Y"":2},{""X"":3,""Y"":1}]","[{""Idx"":0,""Rot"":false},{""Idx"":3,""Rot"":false}",""
```

## Example output
Each worker has it's own set of output files. A file with found solutions in *_[worker_id].solutions.csv and a file tracking puzzle status and metadata in *_[worker_id].status.csv.


*.solutions.csv example:
* ```puzzle_id``` and ```job_id``` link back to
* ```tiles``` contains the solutions as a json encoded array of tile objects with the width and height of the tile in ```W``` and ```H```, the coordinates of the lower left tile corner in ```X``` and ```Y``` and the tile rotation in ```T```
* ```hash``` is a sha1 hash of the tiles field
```
puzzle_id,job_id,tiles,tiles_hash
44,44,"[{""W"":16,""H"":4,""X"":0,""Y"":0,""T"":false},{""W"":16,""H"":3,""X"":0,""Y"":4,""T"":false},{""W"":16,""H"":1,""X"":0,""Y"":7,""T"":false},{""W"":16,""H"":1,""X"":0,""Y"":8,""T"":false},{""W"":16,""H"":1,""X"":0,""Y"":9,""T"":false},{""W"":14,""H"":2,""X"":0,""Y"":10,""T"":false},{""W"":12,""H"":2,""X"":0,""Y"":12,""T"":false},{""W"":10,""H"":1,""X"":16,""Y"":0,""T"":true},{""W"":6,""H"":1,""X"":12,""Y"":12,""T"":false},{""W"":5,""H"":1,""X"":12,""Y"":13,""T"":false},{""W"":4,""H"":1,""X"":14,""Y"":10,""T"":false},{""W"":3,""H"":1,""X"":14,""Y"":11,""T"":false},{""W"":2,""H"":1,""X"":17,""Y"":0,""T"":true},{""W"":2,""H"":1,""X"":17,""Y"":2,""T"":true},{""W"":2,""H"":1,""X"":17,""Y"":4,""T"":true},{""W"":2,""H"":1,""X"":17,""Y"":6,""T"":true},{""W"":1,""H"":1,""X"":17,""Y"":13,""T"":true},{""W"":1,""H"":1,""X"":17,""Y"":11,""T"":true},{""W"":1,""H"":1,""X"":17,""Y"":8,""T"":true},{""W"":1,""H"":1,""X"":17,""Y"":9,""T"":true}]",b5f4253493e01e7d806a586a941fcaffe22b55f2
46,46,"[{""W"":9,""H"":3,""X"":0,""Y"":0,""T"":true},{""W"":7,""H"":1,""X"":0,""Y"":9,""T"":false},{""W"":6,""H"":2,""X"":3,""Y"":0,""T"":true},{""W"":6,""H"":1,""X"":0,""Y"":10,""T"":false},{""W"":6,""H"":1,""X"":6,""Y"":10,""T"":false},{""W"":5,""H"":4,""X"":5,""Y"":0,""T"":true},{""W"":5,""H"":4,""X"":9,""Y"":0,""T"":true},{""W"":5,""H"":2,""X"":3,""Y"":6,""T"":false},{""W"":5,""H"":2,""X"":8,""Y"":6,""T"":false},{""W"":5,""H"":1,""X"":3,""Y"":8,""T"":false},{""W"":5,""H"":1,""X"":7,""Y"":9,""T"":false},{""W"":5,""H"":1,""X"":5,""Y"":5,""T"":false},{""W"":4,""H"":2,""X"":13,""Y"":0,""T"":true},{""W"":4,""H"":2,""X"":13,""Y"":6,""T"":true},{""W"":4,""H"":1,""X"":8,""Y"":8,""T"":false},{""W"":4,""H"":1,""X"":10,""Y"":5,""T"":false},{""W"":3,""H"":1,""X"":12,""Y"":8,""T"":true},{""W"":2,""H"":1,""X"":13,""Y"":4,""T"":false},{""W"":2,""H"":1,""X"":13,""Y"":10,""T"":false},{""W"":1,""H"":1,""X"":14,""Y"":5,""T"":true}]",6acfa2eda039c486019992bd34a8300eae8d526a
```

*.status.csv example:
* ```puzzle_id``` and ```job_id``` link back to
* ```status```
    * 'solved1' if the solver found 1 solution and the -stop_on_solution option was set to true.
    * 'solved' if the solver finished, either because no solutions were found or -stop_on_solutions was true and all solutions were found.
    * 'interrupted' if the worker was forced to return before finishing the full puzzle or the job "end".
* ```tiles_placed``` describes the number of tiles placed (and possibly removed again) up to this point.
* ```duration``` describes the time taken in nanoseconds for this puzzle or job.
* ```solver_id``` The number in -solver_id as specified when starting the program.
* ```current_state``` The frame configuration at the time of interruption. A json encode array of tiles in the order the solver placed them, ```Idx``` references  a tile index as ordered in ```tiles```, and ```rot``` a boolean, is true if the tile was placed 90 degrees rotated.

```
job_id,puzzle_id,status,tiles_placed,duration,solver_id,current_state
1,1,interrupted,14789,60000466367,20,"[{""Idx"":0,""Rot"":true},{""Idx"":1,""Rot"":true},{""Idx"":2,""Rot"":true},{""Idx"":3,""Rot"":true},{""Idx"":5,""Rot"":true},{""Idx"":7,""Rot"":true},{""Idx"":9,""Rot"":true},{""Idx"":10,""Rot"":true},{""Idx"":11,""Rot"":true},{""Idx"":13,""Rot"":true},{""Idx"":16,""Rot"":false},{""Idx"":12,""Rot"":false}]"
44,44,solved1,20,33111167,20,"[{""Idx"":0,""Rot"":false},{""Idx"":1,""Rot"":false},{""Idx"":2,""Rot"":false},{""Idx"":3,""Rot"":false},{""Idx"":4,""Rot"":false},{""Idx"":5,""Rot"":false},{""Idx"":6,""Rot"":false},{""Idx"":8,""Rot"":false},{""Idx"":9,""Rot"":false},{""Idx"":16,""Rot"":true},{""Idx"":10,""Rot"":false},{""Idx"":11,""Rot"":false},{""Idx"":17,""Rot"":true},{""Idx"":7,""Rot"":true},{""Idx"":12,""Rot"":true},{""Idx"":13,""Rot"":true},{""Idx"":14,""Rot"":true},{""Idx"":15,""Rot"":true},{""Idx"":18,""Rot"":true},{""Idx"":19,""Rot"":true}]"
46,46,solved1,20,20290723,20,"[{""Idx"":0,""Rot"":true},{""Idx"":1,""Rot"":false},{""Idx"":3,""Rot"":false},{""Idx"":4,""Rot"":false},{""Idx"":2,""Rot"":true},{""Idx"":7,""Rot"":false},{""Idx"":9,""Rot"":false},{""Idx"":10,""Rot"":false},{""Idx"":5,""Rot"":true},{""Idx"":11,""Rot"":false},{""Idx"":8,""Rot"":false},{""Idx"":14,""Rot"":false},{""Idx"":16,""Rot"":true},{""Idx"":6,""Rot"":true},{""Idx"":15,""Rot"":false},{""Idx"":12,""Rot"":true},{""Idx"":17,""Rot"":false},{""Idx"":13,""Rot"":true},{""Idx"":18,""Rot"":false},{""Idx"":19,""Rot"":true}]"
1,1,solved,0,530281,20,
15,15,solved,0,332106,20,
```