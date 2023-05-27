package main 

import (
    "hema/hangman/rpc/gameserver"
    "hema/hangman/rpc/playerserver"
    "hema/hangman/rpc/asyncio"
    "fmt"
    "net/rpc"
    "errors"
    "os"
    "net"
    "strings"
    "strconv"
)


//GLOBAL VARIABLE
var io asyncio.IoInstance


type Args struct {}


func JoinGame(ip string, port string, player_server *playerserver.HangmanPlayerServer) error{
    
    fmt.Println("Called Join")

    //create ones own player server
    //player_server, err := playerserver.Init(io)
    
    fmt.Println("created player server")


    //connect to game server

    client, err := rpc.Dial("tcp", ip + ":" + port)
    //client, err := rpc.DialHTTP("tcp", ip + ":" + port)

    if(err != nil){
        fmt.Println(err)
        return errors.New("Could not connect to Game Server")
    }else{
        fmt.Println("Connected!")
    }

    

    //request to join game 
    var addr gameserver.PlayerAddress = player_server.Address



    var pid gameserver.PlayerId
    err = client.Call("HangmanGameServer.Join", &addr, &pid)

    if err != nil {
        fmt.Println(err)
        return errors.New("Failed to join game")
    }

    player_server.Pid = pid
    player_server.GameClient = client
    
    return nil

}


func ChooseGame(player *playerserver.HangmanPlayerServer) error{


    var ip, port string


    fmt.Print("Enter IP Address of game host: ")
    
    input := io.RequestLine()

    ip = input.Result
    
    fmt.Print("Enter Port of game host: ")
    input = io.RequestLine()
    port = input.Result

    err := JoinGame(ip, port, player)
    
    if err != nil {
        return err
    }

    return nil
}

func host(player *playerserver.HangmanPlayerServer) error {

    

    //create game server
    _, err := gameserver.Init()
    //game_server, err := gameserver.Init()

    //rpc.HandleHTTP()

    if(err != nil){
        return  err
    }

    //fmt.Println(gameserver.GetPlayerCount())

    //Join the game as player
    err = JoinGame("127.0.0.1", "1234", player)

    if err != nil{
        fmt.Println(err)
        return err
    }else{
        fmt.Println("Joined!")
    }

    return nil

}

func WaitingForGameToEnd(player *playerserver.HangmanPlayerServer){
    <- player.GameEnd

    fmt.Println("Finished Waiting!")
}

func lobby(role int, player *playerserver.HangmanPlayerServer)(error){

    var choice int
/*
    fmt.Println("----- HANGMAN LOBBY -----\n")

    ip := "1.2.3.4"
    port := "1111"
    
    fmt.Println("IP:\t", ip, "\nPort:\t", port);

    fmt.Println("\n==========================\n")

    var playerString string = "PlayerX"

    fmt.Println("PLAYERS [6/8]\n")

    fmt.Println(playerString, "\t", playerString)
    fmt.Println(playerString, "\t", playerString)
    fmt.Println(playerString, "\t", "")
    fmt.Println(playerString, "\t", "")
    fmt.Println("\n==========================\n\n")
    fmt.Println("\n==========================\n")
    fmt.Println("[1] Start Game (Host only)\n[2] Leave Lobby\n")
*/

    var res bool
    var err error



    input := io.RequestLineSignal(player.GameStart)

    //close(player.GameStart)
        //fmt.Scanf("%d", &choice)

    if(input.Canceled == true){
        WaitingForGameToEnd(player)
    }else{
    
        choice, err = strconv.Atoi(input.Result)

        if err != nil {
            fmt.Println(err)
        }

        if choice == 1 {

            if role == 1 {
                go player.GameClient.Call("HangmanGameServer.StartGame", &player.Pid, &res)
                if(err != nil){ 
                    fmt.Println("Error while starting game")
                    return err
                }else{
                    WaitingForGameToEnd(player)
                }
            }         


        }else if choice == 2 {

            err = player.QuitLobby()
            if(err != nil){
                fmt.Println("Error while quitting game")
                return err
            }
            os.Exit(1);

        }
    }

    return nil

}

func findPort()(int, error){

   addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")

    if err != nil {
        fmt.Println("Error:", err)
        return -1, err
    }

   list, err :=net.ListenTCP("tcp", addr)

    if err != nil {
        fmt.Println("Error:", err)
        return -1, err
    }

   addr2 := list.Addr().String()

    err = list.Close()
    if err != nil {
        fmt.Println("Error:", err)
        return -1, err
    }

   idx := strings.Index(addr2, ":")

    port, err := strconv.Atoi(addr2[idx+1:])
    if err != nil {
        fmt.Println("Error:", err)
        return -1, err
    }
   fmt.Println("Port:", port)

    return port, nil

}


func test(){

    io, _ := asyncio.NewIoInstance()

    io.Launch()
    
    for{
    
        resp := io.RequestLineTimeout(10)

        if resp.Canceled == true{
            fmt.Println("Timed out")
        }else{
            fmt.Println("User Input:", resp.Result)
        }

    }

}

func main() {


    //test()
    //return

    io, _ = asyncio.NewIoInstance()
    io.Launch()


    var choice int
    var player *playerserver.HangmanPlayerServer
    var err error

    //Create PlayerServer once at the beginning
    player, err = playerserver.Init(io)

    for{
        fmt.Println("Welcome To HANGMAN!\n")

        fmt.Println("[1] Host New Game\t\t[2] Join Game")
        input := io.RequestLine()

        choice, err = strconv.Atoi(input.Result)

        if err != nil {
            fmt.Println(err, "Choose between '1' and '2'")
            continue
        }

        if choice == 1{
            err  = host(player)
        } else if choice == 2{
            err  = ChooseGame(player)
        }

        if(err == nil){
            lobby(choice, player)
        }
    }



}

