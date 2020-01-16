// Copyright 2020 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package log

import (
	"log"
)

var verbose bool

func Init(verbose bool) {
	verbose = verbose
}

func Error(fmt string, args ...string) {
	log.Printf("[ERROR] "+fmt, args)
}

func Info(fmt string, args ...string) {
	log.Printf("[INFO] "+fmt, args)
}

func Verbose(fmt string, args ...string) {
	if verbose {
		log.Printf("[VERBOSE] "+fmt, args)
	}
}
