package tcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"node/database"
)

var Port string
var Nodes map[string]net.Conn
var Listener net.Listener

func Run(httpPort string) {
	fmt.Printf("TCP Port : %s\n", Port)

	Nodes = make(map[string]net.Conn)
	var err error
	Listener, err = net.Listen("tcp", ":"+Port)
	if err != nil {
		fmt.Printf("TCP Listen 실패 : %s\n", err)
		os.Exit(1)
	}

	go func() {
		for {
			conn, err := Listener.Accept()
			if err != nil {
				fmt.Printf("신규 노드 연결 승인 실패 : %s\n", err)
				continue
			}

			buffer := make([]byte, 512)
			bytesRead, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("신규 노드 포트 수신 실패 : %s\n", err)
				continue
			}
			newNodePort := string(buffer[:bytesRead])

			fmt.Printf("노드(%s)와 연결\n", newNodePort)
			Nodes[newNodePort] = conn

			go HandleNode(conn, newNodePort)
		}
	}()

	JoinP2P(httpPort)
}

func HandleNode(conn net.Conn, remotePort string) {
	defer func() {
		conn.Close()
		delete(Nodes, remotePort)
	}()

	for {
		// 파일 이름 수신
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("노드(%s) 연결 끊김\n", remotePort)
			return
		}
		fileName := string(buffer[:bytesRead])
		fmt.Println("파일 이름 : ", fileName)

		// 파일 크기 수신
		buffer = make([]byte, 1024)
		_, err = conn.Read(buffer)
		if err != nil {
			fmt.Printf("파일 크기 수신 실패:%s\n", err)
			return
		}
		fileSize, err := strconv.ParseInt(string(bytes.Trim(buffer, "\x00")), 10, 64)
		if err != nil {
			fmt.Printf("파일 크기 파싱 실패:%s\n", err)
			return
		}

		// 파일 수신
		uploadDir := "./uploads/" + database.DbName + "/"
		os.MkdirAll(uploadDir, os.ModePerm)
		filePath := uploadDir + fileName

		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return
		}
		defer file.Close()

		_, err = io.CopyN(file, conn, fileSize)
		if err != nil {
			fmt.Printf("Error saving file content: %v\n", err)
			return
		}

		err = database.SaveFileToDB(fileName, filePath)
		if err != nil {
			fmt.Printf("database saveing fail : %v\n", err)
			return
		}

		fmt.Printf("File '%s' saved\n", fileName)
	}
}

func JoinP2P(httpPort string) {
	//중앙에 nodeList 요청
	values := url.Values{}
	values.Add("port", httpPort)

	url := "http://localhost:8080/?" + values.Encode()

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("노드 리스트 받아오기 실패:%s\n", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	var nodeList []string
	err = json.NewDecoder(response.Body).Decode(&nodeList)
	if err != nil {
		fmt.Printf("노드 리스트 디코딩 실패:%s\n", err)
		return
	}

	//nodeList에 각각 dial
	for _, nodePort := range nodeList {
		nodeTCPPort, err := ConvertHTTPToTCPPort(nodePort)
		if err != nil {
			fmt.Printf("HTTP TO TCP 포트변환 실패:%s\n", err)
			continue
		}
		conn, err := net.Dial("tcp", ":"+nodeTCPPort)
		if err != nil {
			fmt.Printf("노드(%s)에 연결 실패:%s\n", nodePort, err)
			continue
		}

		fmt.Printf("노드(%s)와 연결\n", nodePort)
		Nodes[nodePort] = conn
		go HandleNode(conn, nodePort)

		//연결된 노드에 본인 포트 전달
		conn.Write([]byte(httpPort))
	}

	fmt.Printf("P2P 네트워크 입장 성공\n")
}

func ConvertHTTPToTCPPort(HttpPort string) (string, error) {
	port, err := strconv.Atoi(HttpPort)
	if err != nil {
		return "", err
	}
	port += 1000
	TcpPort := strconv.Itoa(port)
	return TcpPort, nil
}

func SendFileToOtherNodes(file multipart.File, fileName string, filePath string) {
	for nodeName, conn := range Nodes {
		conn.Write([]byte(fileName))

		stat, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Error getting file info: %v\n", err)
			continue
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)
		fmt.Println("파일 사이즈 전달 : ", fileSize)
		conn.Write([]byte(fileSize))

		_, err = io.Copy(conn, file)
		if err != nil {
			fmt.Printf("Error sending file content to %s: %v", nodeName, err)
			return
		}

		fmt.Printf("File sent to node %s\n", nodeName)
	}
}
