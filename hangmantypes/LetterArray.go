package hangmantypes

import(
    "fmt"
    "errors"
)



type LetterArray struct{
    Letters string 
    Status [26] uint8
}

func (l_arr *LetterArray)Init(){
    l_arr.Letters = "abcdefghijklmnopqrstuvwxyz"
}

func (l_arr *LetterArray)Update(pos uint8, found bool) error{


    if(pos < 'a' || pos > 'z'){
        return errors.New("LetterArray.update: Tried to update with a character not in range [a, z]")
    }

    if found == true{
        l_arr.Status[pos-'a'] = 1
    }else{
        l_arr.Status[pos-'a'] = 2
    }

    return nil
}

func (l_arr LetterArray)PrintArr(){
    for pos, char := range l_arr.Letters{
        if(l_arr.Status[pos] == 0){
            fmt.Print(string(char))
        }else{
            fmt.Print("-")
        }
    }

    fmt.Println("")
}
