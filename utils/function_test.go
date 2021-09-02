/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package utils

import "testing"

func BenchmarkV1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GetRandomString(16) // run fib(30) b.N times
	}
}
func BenchmarkV2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		//GetRandomStringV2(16) // run fib(30) b.N times
	}
}
