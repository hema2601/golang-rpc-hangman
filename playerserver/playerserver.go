package playerserver

import(
    "errors"
    "net"
    //"net/http"
    "net/rpc"
    "hema/hangman/rpc/gameserver"
    "fmt"
    "hema/hangman/rpc/hangmantypes"
    "hema/hangman/rpc/hangmandisplay"
    "strings"
    "strconv"
    "hema/hangman/rpc/asyncio"
)


//type GameClient *rpc.Client

type HangmanPlayerServer struct {
    Pid gameserver.PlayerId
    //Game_Client GameClient
    GameClient *rpc.Client
    GameStart chan int
    GameEnd chan int
    Address gameserver.PlayerAddress
    io asyncio.IoInstance
}

func (h_playerserver *HangmanPlayerServer) GetString( pid *gameserver.PlayerId, reply *asyncio.IoResponse) error{

    if(*pid == h_playerserver.Pid){
        fmt.Println("Choose the String for this round! You have 20 seconds")
       
        *reply = h_playerserver.io.RequestLineTimeout(20) 

    }else{
        fmt.Println("Waiting for Player", *pid, "to choose this round's string")
        
    }
    return nil
}

func (h_playerserver *HangmanPlayerServer) EndGame( pid *gameserver.PlayerId, reply *asyncio.IoResponse) error{

    fmt.Println("The game has ended. Send 'q' if you want to leave, send anything else if you want to continue. You have 10 seconds")

  //  if(*pid == h_playerserver.Pid){
    //    fmt.Println("Choose the String for this round! You have 20 seconds")
       
    *reply = h_playerserver.io.RequestLineTimeout(10) 

    if reply.Canceled == true || reply.Result == "q"{
        fmt.Println("Quit Game")
        h_playerserver.GameEnd <- 1
    }else{
        fmt.Println("Continue")
    }

    //}else{
    //    fmt.Println("Waiting for Player", *pid, "to choose this round's string")
        
    //}

    fmt.Println("Returning")
    return nil
}

func (h_playerserver *HangmanPlayerServer) DoTurn( pid *gameserver.PlayerId, reply *asyncio.IoResponse) error{

    if(*pid == h_playerserver.Pid){
        fmt.Println("Your Turn! You have 20 seconds")

        *reply = h_playerserver.io.RequestLineTimeout(20)



    }else{
        fmt.Println("Waiting for Player", *pid, "to make a guess")

    }
    return nil
}
func (h_playerserver *HangmanPlayerServer) UpdateDisplay(h_game *hangmantypes.HangmanGame, reply *bool) error{


    h_game.PrintState();
    fmt.Println("Updated Display...")

    return nil

}

func (h_playerserver *HangmanPlayerServer) StartGame(args *uint8, reply *uint8) error{

    //go h_playerserver.InteractiveInput()

    fmt.Println("The game has started!")
    
    h_playerserver.GameStart <- 1

    return nil
}




func (p HangmanPlayerServer)QuitLobby()(error){
    
    var res bool

    err := p.GameClient.Call("HangmanGameServer.QuitLobby", &p.Pid, &res)

    if(err != nil){
        fmt.Println("Error while quitting game")
        return err
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

func (h_playerserver *HangmanPlayerServer)UpdateLobby(server *gameserver.PlayerMetadata, reply *uint8) error{

    hangmandisplay.Clear()


    fmt.Println("----- HANGMAN LOBBY -----\n")

    ip := "1.2.3.4"
    port := "1111"

    fmt.Println("IP:\t", ip, "\nPort:\t", port);

    fmt.Println("\n==========================\n")

    fmt.Println("PLAYERS [", server.PlayerCount, "/ 8 ]\n")
    
    for i:= 0; i < 4; i++ {

        if(i >= server.PlayerCount){
            break
        }

        fmt.Print("Player", server.Players[i])
        if i + 4 >= server.PlayerCount {
            fmt.Println()
        }else{
            fmt.Print("\tPlayer", server.Players[i*2], "\n")

        }

    }
    fmt.Println("\n==========================\n\n")
    fmt.Println("\n==========================\n")
    fmt.Println("[1] Start Game (Host only)\n[2] Leave Lobby\n")


    return nil

}

func Init(io asyncio.IoInstance) (*HangmanPlayerServer, error){

    h_playerserver :=  new(HangmanPlayerServer)
   
    handler := rpc.NewServer()

    err := handler.Register(h_playerserver)

    if err != nil{
        fmt.Println(err)
        return h_playerserver, err
    }

    handler.HandleHTTP("/HangmanPlayerServer", "/debug/HangmanPlayerServer")

    newPort, err := findPort()
    if err != nil{
        fmt.Println(err)
        return h_playerserver, err
    }

    port := strconv.Itoa(newPort)

    listener, err := net.Listen("tcp", "0.0.0.0:" + port)

    if err != nil{
        return h_playerserver, errors.New("Failed to Listen")
    }

    h_playerserver.Address.Ip = "0.0.0.0"
    h_playerserver.Address.Port = port
    h_playerserver.GameStart = make(chan int, 1)
    h_playerserver.GameEnd = make(chan int, 1)
    h_playerserver.io = io

  //  go http.Serve(listener, nil)

     go func() {
         for{
             fmt.Println("WAiting", port)
             cxn, _ := listener.Accept()
             fmt.Println("ACcepted", port)
             go handler.ServeConn(cxn)
         }
     }()



    return h_playerserver, nil


}
