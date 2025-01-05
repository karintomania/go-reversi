# Go Reversi
CLI application to play reversi in terminal.   
![screen recording](https://github.com/user-attachments/assets/ec7d106e-3daf-41a5-995a-c1cdc0e7bf05)

# How to play

## Docker (Recommened)
This command starts a game in single player mode with 8x8 board.  
```
docker run --rm -it ghcr.io/karintomania/go-reversi:latest 
```

See Examples section for more options.  

## Download Binary
Install the binary and run the file from [Releases Page](https://github.com/karintomania/go-reversi/releases).  

# Examples
These are the options you can specify:
```
  -h show help
  -n int
        Dimension of the board. (Default: 8) (default 8)
  -p int
        1 for Single Play, 2 for 2 Players. (Default: 1) (default 1)
```

Single Player mode with 8x8 board. (Change the command according to your platform and version)
```
./go-reversi-0.1-linux-x86
```

2 Players mode with 6x6 board.
```
./go-reversi-0.1-linux-x86 -n 6 -p 2
```


You can watch 2 AIs playing if you want.
```
./go-reversi-0.1-linux-x86 -p 0
```
