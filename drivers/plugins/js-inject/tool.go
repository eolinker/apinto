package js_inject

import (
	"bytes"
	"fmt"

	"golang.org/x/net/html"
)

// injectJavaScript将JavaScript代码插入到HTML中
func injectJavaScript(originalHTML, jsCode string) (string, error) {
	// 解析 HTML 字符串
	doc, err := html.Parse(bytes.NewReader([]byte(originalHTML)))
	if err != nil {
		return "", fmt.Errorf("parse html error: %w", err)
	}
	node := findHead(doc)
	if node == nil {
		// 在每个 HTML 标签后添加 <head> 元素
		addHeadAfterTags(doc)
		node = findHead(doc)
	}
	insertCodeInHead(node, jsCode)
	// 将修改后的 HTML 打印出来
	var modifiedHTML bytes.Buffer
	if err := html.Render(&modifiedHTML, doc); err != nil {
		return "", fmt.Errorf("render html error: %w", err)
	}

	return modifiedHTML.String(), nil

}

func findHead(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "head" {
		return node
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findHead(child); found != nil {
			return found
		}
	}

	return nil
}

func addHeadAfterTags(node *html.Node) {
	if node.Type == html.ElementNode {
		// 创建 <head> 元素
		newHead := &html.Node{
			Type:     html.ElementNode,
			DataAtom: 0,
			Data:     "head",
		}
		// 在当前元素后插入 <head> 元素
		node.Parent.InsertBefore(newHead, node)
		// 递归处理子节点
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			addHeadAfterTags(child)
		}
	}

	// 递归处理下一个兄弟节点
	if node.NextSibling != nil {
		addHeadAfterTags(node.NextSibling)
	}
}

func insertCodeInHead(headNode *html.Node, jsCode string) {
	// 创建一个新的 <script> 元素（示例代码）
	newScript := &html.Node{
		Type:     html.ElementNode,
		Data:     "script",
		DataAtom: 0,
	}

	// 在 <script> 中插入要添加的代码
	newScript.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: jsCode,
	})

	// 将新的 <script> 元素插入到 <head> 的第一个子标签位置
	if headNode.LastChild != nil {
		headNode.InsertBefore(newScript, headNode.FirstChild)
	} else {
		headNode.AppendChild(newScript)
	}
}
