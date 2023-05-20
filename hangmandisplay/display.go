package hangmandisplay

import(
    "fmt"
    "bufio"
    "os"
    "os/exec"
    "time"
)

func Clear(){
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}

func PrintStart(){
    fmt.Print("Starting your game")

    i := 0

    for i < 3{
        time.Sleep(time.Millisecond * 400)
        fmt.Print(".")
        i++

    }
       
    time.Sleep(time.Millisecond * 700)

    fmt.Println("");
}

func PrintCharByChar(milli time.Duration, str string){

    for _, char := range str{
        fmt.Print(string(char))
        time.Sleep(time.Millisecond * milli)
    }
    fmt.Println("")

}

func ScanningStdout(quit chan bool){
    scanner := bufio.NewScanner(os.Stdout)
    
    for{
        select {
            case <- quit:
                return
            default:
                scanner.Scan()
                fmt.Println("Scanned:", scanner.Text())
        }
    }
}

