package util

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"fmt"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func GetIPv6Address(decstr []string)(ip string){
	//Build Unsummarized IPv6 Address
	for i := 0; i < len(decstr); i+=2{
		for j := 0; j < 2; j++{
			dec, _ := strconv.Atoi(decstr[i+j])
			ip += fmt.Sprintf("%02x", dec)
		}
		if i < len(decstr) - 2{
			ip += ":"
		}
	}

	//Summarize Adjacent Zero Octets
	split := strings.Split(ip, ":")
	maxAdjacentIdx := []int{0, 0}
	curAdjacentIdx := []int{0, 0}
	maxAdjacent := 0
	curAdjacent := 0
	for i := 0; i < len(split); i++{
		//Trim Leading Zeros
		for j := 0; j < len(split[i]); j++{
			//Different Than Zero, Slice From Current Index To Last
			if split[i][j] != '0'{
				split[i] = split[i][j:]
				break
			}
			//Found All Zeros, Make It A Single Zero
			if j == len(split[i])-1 && split[i][j] == '0'{
				split[i] = "0"
			}
		}
		//Check For Zeros
		if split[i] == "0"{
			if curAdjacent == 0{
				curAdjacentIdx[0] = i //Start Counting Indexes
			}
			curAdjacent++ //Increment Adjacency Counter
		}
		if (split[i] != "0" || i == len(split) - 1) && curAdjacent > 0{
			curAdjacentIdx[1] = i //Stop Counting Indexes

			//Count The Last Zero
			if split[i] == "0" && i == len(split) - 1{
				curAdjacentIdx[1] = i+1
			}

			//If Current Adjacency Counter Is Larger Than Maximum Counter Value Set New Maximum
			if curAdjacent >= maxAdjacent{
				maxAdjacent = curAdjacent
				maxAdjacentIdx = curAdjacentIdx
			}

			//Reset Current Adjacency Counter And Indexes
			curAdjacentIdx = []int{0, 0}
			curAdjacent = 0
		}
	}
	//Adjacency Indexes' Difference Being Bigger Than 1, Means That There Was More Than Just One Adjacent Zero
	if maxAdjacentIdx[1] - maxAdjacentIdx[0] > 1{
		ip = strings.Join(split[0:maxAdjacentIdx[0]], ":") + "::" + strings.Join(split[maxAdjacentIdx[1]:], ":")
	}

	return
}