// Copyright 2016 The go-ethereum Authors
// Copyright 2018 The go-etherzero Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package masternode

//import (
//	"os"
//	"os/user"
//	"path/filepath"
//	"runtime"
//
//)
//
//const (
//	DefaultHTTPHost = "localhost" // Default host interface for the HTTP RPC server
//	DefaultHTTPPort = 1646        // Default TCP port for the HTTP RPC server
//	DefaultWSHost   = "localhost" // Default host interface for the websocket RPC server
//	DefaultWSPort   = 2546        // Default TCP port for the websocket RPC server
//)
//
//// DefaultConfig contains reasonable default settings.
//var DefaultConfig = Config{
//	DataDir:          DefaultDataDir(),
//	HTTPPort:         DefaultHTTPPort,
//	HTTPModules:      []string{"net", "web3"},
//	HTTPVirtualHosts: []string{"localhost"},
//	WSPort:           DefaultWSPort,
//	WSModules:        []string{"net", "web3"},
//}
//
//// DefaultDataDir is the default data directory to use for the databases and other
//// persistence requirements.
//func DefaultDataDir() string {
//	// Try to place the data folder in the user's home dir
//	home := homeDir()
//	if home != "" {
//		if runtime.GOOS == "darwin" {
//			return filepath.Join(home, "Library", "Etzmasternode")
//		} else if runtime.GOOS == "windows" {
//			return filepath.Join(home, "AppData", "Roaming", "Ethzmasternode")
//		} else {
//			return filepath.Join(home, ".ethzmasternode")
//		}
//	}
//	// As we cannot guess a stable location, return empty and handle later
//	return ""
//}
//
//func homeDir() string {
//	if home := os.Getenv("HOME"); home != "" {
//		return home
//	}
//	if usr, err := user.Current(); err == nil {
//		return usr.HomeDir
//	}
//	return ""
//}
//
