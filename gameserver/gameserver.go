package gameserver

import(
    "hema/hangman/rpc/hangmantypes"
    "errors"
    "net"
    "net/rpc"
    "fmt"
    "strconv"
    "hema/hangman/rpc/asyncio"
)


type PlayerId uint16

type PlayerAddress struct{
    Ip string
    Port string
}


type Player struct{
    Pid PlayerId
    IsHost bool
    Addr PlayerAddress
    Client *rpc.Client
    //Client PlayerClient
}

func InitPlayer(id PlayerId, addr PlayerAddress)( Player, error){
    
    var player Player
    
    client, err := rpc.Dial("tcp", addr.Ip + ":" + addr.Port)

    if err != nil{
        fmt.Println(err)
        return player, errors.New("Could not connect to Client")
    }
    player.Pid = id
    player.IsHost = (id == 0)
    player.Addr = addr
    player.Client = client

    return player, nil
}

type LobbyData struct{
    PlayerCount int
    Players [] PlayerId
    Addr PlayerAddress
}

type HangmanGameServer struct {
    H_game hangmantypes.HangmanGame
    Addr PlayerAddress
    PlayerCount int
    Players [] Player
    LastTurn PlayerId
    CurrPlayer PlayerId
    Running bool
    Active bool
}

func (h_gameserver *HangmanGameServer) GetNewPid() (PlayerId, error){

   for i := 0; i < h_gameserver.PlayerCount; i++ {
        _, err := h_gameserver.GetPlayerByPid(PlayerId(i))
        
        if err != nil{
            return PlayerId(i), nil
        }

   }

    return PlayerId(h_gameserver.PlayerCount), nil

}

func (h_gameserver *HangmanGameServer) GetPlayerByPid(pid PlayerId) (*Player, error){

    for _, p := range h_gameserver.Players{
        if p.Pid == pid{
            return &p, nil
        }
    }

    return nil, errors.New("GetPlayerByBid: No player with pid" + strconv.Itoa(int(pid)))
}
func (h_gameserver *HangmanGameServer) GetIndexByPid(pid PlayerId) (int, error){

    for i,p  := range h_gameserver.Players{
        if p.Pid == pid{
            return i, nil
        }
    }

    return -1, errors.New("GetIndexByPid: No player with pid" + strconv.Itoa(int(pid)))
}

func (h_gameserver *HangmanGameServer)Join(addr *PlayerAddress, pid *PlayerId) error {
   
    var err error

    if h_gameserver.Running == true{
        return errors.New("Game is running")
    }
    
    if h_gameserver.Active == false{
        return errors.New("Inactive server")
    }

    *pid, err = h_gameserver.GetNewPid()  
    //create new player Client
    player, err := InitPlayer(*pid, *addr)

    if(err != nil){
        fmt.Println(err)
        return err
    }
    
    //add the player to the list on success
    h_gameserver.Players = append(h_gameserver.Players, player)

    //register player id as return value
    h_gameserver.PlayerCount++

    var tmp uint8
    var player_info LobbyData
    player_info.PlayerCount = h_gameserver.PlayerCount
    player_info.Players = make([]PlayerId, player_info.PlayerCount)
    player_info.Addr = h_gameserver.Addr 

    for i, p := range h_gameserver.Players{
        player_info.Players[i] = p.Pid
    }

    for _, p := range h_gameserver.Players{
        //fmt.Println("Hi?")
        err = p.Client.Call("HangmanPlayerServer.UpdateLobby", &player_info, &tmp)
    
        if err != nil{
            fmt.Println(err)
        }
    
    }

    //fmt.Println("A new player joined! (Player", *pid, ")")

    return nil

}

func (h_gameserver *HangmanGameServer)QuitLobby(pid *PlayerId, result *bool) error{

    //fmt.Println("Player", *pid, "quit the lobby")
    
    p, _ := h_gameserver.GetPlayerByPid(*pid)


    h_gameserver.Quit(pid, result)

    if p.IsHost == true{
        h_gameserver.Active = false
        fmt.Println("Admin Left!")

        var tmp1, tmp2 bool

        for _, p := range h_gameserver.Players{
            err := p.Client.Call("HangmanPlayerServer.TerminateLobby", &tmp1, &tmp2)
            if err != nil {
                fmt.Println("Error:", err)
            }
        }

        //[TODO] Quit all other players

        return nil;

    }

    

    var tmp uint8
    var err error
    var player_info LobbyData

    player_info.PlayerCount = h_gameserver.PlayerCount
    player_info.Players = make([]PlayerId, player_info.PlayerCount)
    player_info.Addr = h_gameserver.Addr 

    for i, p := range h_gameserver.Players{
        player_info.Players[i] = p.Pid
    }
    for _, p := range h_gameserver.Players{
        //fmt.Println("Updating!")
        err = p.Client.Call("HangmanPlayerServer.UpdateLobby", &player_info, &tmp)

        if err != nil{
            fmt.Println(err)
        }
    }

    return nil
}


func (h_gameserver *HangmanGameServer)CleanupServer() error{
    return nil
}

func (h_gameserver *HangmanGameServer)Quit(pid *PlayerId, result *bool) error{

  //  fmt.Println("Player", *pid, "quit the lobby")

    var p *Player = nil

    p, err := h_gameserver.GetPlayerByPid(*pid)

    if(err != nil){
        return err
    }

    err = p.Client.Close()
    if err != nil {
        return err
    }
    
    for i, player := range h_gameserver.Players{
        if(i != h_gameserver.PlayerCount-1 && player.Pid == *pid){
            h_gameserver.Players[i] = h_gameserver.Players[h_gameserver.PlayerCount-1]
            break
        }
    }
    h_gameserver.PlayerCount--
    //fmt.Println("Decremented player count to ", h_gameserver.PlayerCount)

    h_gameserver.Players = h_gameserver.Players[:h_gameserver.PlayerCount]
/*
    var tmp uint8


    var player_info PlayerMetadata

    player_info.PlayerCount = h_gameserver.PlayerCount
    player_info.Players = make([]PlayerId, player_info.PlayerCount)

    for i, p := range h_gameserver.Players{
        player_info.Players[i] = p.Pid
    }
    for _, p := range h_gameserver.Players{
        fmt.Println("Updating!")
        err = p.Client.Call("HangmanPlayerServer.UpdateLobby", &player_info, &tmp)

        if err != nil{
            fmt.Println(err)
        }
    }
*/
    return nil
}

func (h_gameserver *HangmanGameServer)StartGame(pid *PlayerId, result *bool) error {
    

    var tmp1, tmp2 uint8 

    h_gameserver.Running = true 

    for _, p := range h_gameserver.Players{
        p.Client.Call("HangmanPlayerServer.StartGame", &tmp1, &tmp2)
    }



//    h_gameserver.H_game.Init()

    h_gameserver.Run()

    return nil
}

func EndGamePlayerChoice(p Player, pid PlayerId, done chan PlayerId, error_chan chan error, res *asyncio.IoResponse){
    
    //fmt.Println("Send msg to ", pid)
    err := p.Client.Call("HangmanPlayerServer.EndGameChoice", &pid, res)
    if err != nil{
        fmt.Println("Error in EndGamePlayerChoice:", err)
    }
    
    //fmt.Println("Received answer from ", pid)

    done <- pid
    error_chan <- err

    return 

}


func(h_gameserver *HangmanGameServer) SendTerminate(reason int, p Player, done chan bool) error{

    var res bool

    err := p.Client.Call("HangmanPlayerServer.EndGame", &reason, &res)

    if err != nil{
        fmt.Println("Error in SendTerminate:", err)
    }

    done <- true

    return err

}
func(h_gameserver *HangmanGameServer) TerminatePlayers(fewPlayers bool, noAdmin bool) error{

    var tmp bool
   
    done := make(chan bool)

    var reason int
    if noAdmin == true {
        reason = 1
    }else if fewPlayers == true {
        reason = 2
    }else{
        //fmt.Println("Returning Prematurely")
        return nil
    }

    //fmt.Println("Number of Players Left:", h_gameserver.PlayerCount)

    for _, p := range h_gameserver.Players{
        //fmt.Println("Sending EndGame to Player", p.Pid)

        go h_gameserver.SendTerminate(reason, p, done)
    }

    for _,_ = range h_gameserver.Players{
        <-done
    }

    players_copy := h_gameserver.Players[:]

    for _, p := range players_copy{
        //fmt.Println("Sending Quit to Player", p.Pid)
        h_gameserver.Quit(&p.Pid, &tmp);
    }

    return nil

}

func (h_gameserver *HangmanGameServer)GameEpilogue() (bool, error) {
    

    res := make([]asyncio.IoResponse, h_gameserver.PlayerCount)
    quit := make([]PlayerId, 0)
    done := make(chan PlayerId)
    error_chan := make(chan error)



    for i, p := range h_gameserver.Players{
        go EndGamePlayerChoice(p, p.Pid, done, error_chan, &res[i]) 
    }
    
    var idx int
    var curr_pid PlayerId
    var temp bool
    var err error
    adminLeft := false

    count := h_gameserver.PlayerCount

    //fmt.Println(count, " players left at the beginning")

    for i := 0; i < count; i++{
        //fmt.Println("Waiting for player response")
         curr_pid = <- done;
        //fmt.Println("Response from Player", curr_pid)

         idx, err = h_gameserver.GetIndexByPid(curr_pid);
         
         err = <- error_chan
         if err != nil{
            fmt.Println("Error in GameEpilogue:", err)
            return false, err
         }
    

         if res[idx].Canceled == true {
            if h_gameserver.Players[idx].IsHost == true{
                    //fmt.Println("Admin left!")
                    adminLeft = true

            }
            //fmt.Println("Player", curr_pid, "quit (Cancelled)")
            quit = append(quit, curr_pid)
            //h_gameserver.Quit(&curr_pid, &temp);
         }else{
            if res[idx].Result == "q"{
                if h_gameserver.Players[idx].IsHost == true{
                    //fmt.Println("Admin left!")
                    adminLeft = true
                }
                //fmt.Println("Player", curr_pid, "quit (res[",idx,"]", res[idx].Result, ")")
                //h_gameserver.Quit(&curr_pid, &temp);
                quit = append(quit, curr_pid)
            }

            

         }


    }

    //fmt.Println("Got all responses");

    for _, id := range quit {
        h_gameserver.Quit(&id, &temp);
    }



    if(h_gameserver.PlayerCount < 2 || adminLeft == true){
        //fmt.Println("Quitting Game")
        h_gameserver.TerminatePlayers(h_gameserver.PlayerCount < 2, adminLeft == true)
        h_gameserver.Running = false
    }



    return h_gameserver.Running, nil
}

func GetPlayerString(p Player, pid PlayerId, done chan PlayerId, error_chan chan error, res *asyncio.IoResponse){
    err :=p.Client.Call("HangmanPlayerServer.GetString", &pid, res)
    if err != nil{
        fmt.Println("Error in GetPlayerString:", err)
    }

    done <- p.Pid
    error_chan <- err

    return 

}

func (h_gameserver *HangmanGameServer) ChooseString() (asyncio.IoResponse, error){

    pid := h_gameserver.LastTurn

    fake_res := make([]asyncio.IoResponse, h_gameserver.PlayerCount)
    var res asyncio.IoResponse


    done := make(chan PlayerId)
    error_chan := make(chan error)

    var err error

    for i, p := range h_gameserver.Players{
        if(p.Pid == pid){
            go GetPlayerString(p, pid, done, error_chan, &res) 
        }else{
            go GetPlayerString(p, pid, done, error_chan, &fake_res[i]) 
        }
    }

    for i := 0; i < h_gameserver.PlayerCount; i++{
        if <-done == pid {
            err = <- error_chan
            if(err != nil){
                return res, err 
            } 
            return res, nil
        }
    }

    return res, errors.New("ChooseString(): No Player with this pid")
    

}

func (h_gameserver *HangmanGameServer)Run() error {

    var response asyncio.IoResponse
    var err error

    h_gameserver.LastTurn = PlayerId(0);
    for{
        h_gameserver.H_game.Init()
        h_gameserver.ShareState()

        for{

            response, err = h_gameserver.ChooseString()

            if response.Canceled == false {
                h_gameserver.CurrPlayer = h_gameserver.LastTurn
                h_gameserver.H_game.H_str.Init(response.Result)
                break
            }


            _, _ = h_gameserver.ChooseNextPlayer()

        }

        for !h_gameserver.H_game.Over() {

            h_gameserver.ShareState()

            err = h_gameserver.NextTurn()

            if err != nil{
                fmt.Println("Error during Run:", err)
                return nil
            }

        }

        h_gameserver.ShareState()


        again, _ := h_gameserver.GameEpilogue()

        if again == false {
            //fmt.Println("Ending Game")
            break
        }
    }
    return nil
}


func (s *HangmanGameServer)ChooseNextPlayer() (PlayerId, error){

    
    for i, p := range s.Players{
        if p.Pid == s.LastTurn{
             
            if s.Players[(i+1)%s.PlayerCount].Pid != s.CurrPlayer{
                s.LastTurn = s.Players[(i+1)%s.PlayerCount].Pid
            }else{
                s.LastTurn = s.Players[(i+2)%s.PlayerCount].Pid
            }

            return s.LastTurn, nil
        }
    }


    return PlayerId(9999), errors.New("ChooseNextPlayer: Weird Error Occurred")


}
    
func GetPlayerTurn(p Player, pid PlayerId, done chan PlayerId, error_chan chan error, res *asyncio.IoResponse){
    err :=p.Client.Call("HangmanPlayerServer.DoTurn", &pid, res)
    if err != nil{
        fmt.Println("Error in GetPlayerTurn:", err)
    }

    done <- p.Pid
    error_chan <- err

    return 

}

func (h_gameserver *HangmanGameServer)NextTurn() error{

    for{
        pid, err := h_gameserver.ChooseNextPlayer()

        if err!= nil{
            fmt.Println(err)
            return err
        }

        fake_res := make([]asyncio.IoResponse, h_gameserver.PlayerCount)
        var res asyncio.IoResponse

        done := make(chan PlayerId)
        error_chan := make(chan error)

        for i, p := range h_gameserver.Players{
            if p.Pid == pid {
                go GetPlayerTurn(p, pid, done, error_chan, &res) 
            }else{
                go GetPlayerTurn(p, pid, done, error_chan, &fake_res[i]) 
            }
        }
        for i := 0; i < h_gameserver.PlayerCount; i++{
            if <-done == pid {
                err = <- error_chan
                if(err != nil){
                    return err 
                } 

                if res.Canceled == true {
                    continue
                }

                h_gameserver.H_game.DoTurn(res.Result)
                return nil
            }
        }
    }
return nil



}

func (h_gameserver *HangmanGameServer)ShareState() error {

    res := make([]bool, h_gameserver.PlayerCount)
   

    for i, p := range h_gameserver.Players{
        err := p.Client.Call("HangmanPlayerServer.UpdateDisplay", &h_gameserver.H_game, &res[i])
            if err != nil{
                fmt.Println(err)
                return err
            }
    }

    return nil

}

//Function taken from Stackoverflow
//https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go

func GetOutboundIP() string {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        fmt.Println("Error:", err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP.String()
}

func Init() (*HangmanGameServer, error){

    h_gameserver := new(HangmanGameServer)
    
    h_gameserver.PlayerCount = 0
    h_gameserver.Running = false
    h_gameserver.Active = true
    handler:= rpc.NewServer()

    err := handler.Register(h_gameserver)

    if err != nil{
        fmt.Println(err)
        return h_gameserver, err
    }

    handler.HandleHTTP("/HangmanGameServer", "/debug/HangmanGameServer")

    listener, err := net.Listen("tcp", "0.0.0.0:1234")

    if err != nil{
        
        fmt.Println(err)

        return h_gameserver, errors.New("Failed to Listen")
    }

   h_gameserver.Addr.Ip = GetOutboundIP() 
    h_gameserver.Addr.Port = "1234"


    go func() {
        for{
            cxn, _ := listener.Accept()
            go handler.ServeConn(cxn)
        }
    }()


    return h_gameserver, nil

}
