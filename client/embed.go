// Package client 把客户端运行时嵌进编译器, 生成时拷到应用的 gen/client 目录。
package client

import "embed"

//go:embed runtime.js loader.js idiomorph.esm.js
var FS embed.FS
