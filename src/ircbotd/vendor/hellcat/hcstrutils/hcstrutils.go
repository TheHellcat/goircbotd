package hcstrutils

import (
    "fmt"
)

type HcString struct {
}

func init() {
}

func New() (hcStr *HcString) {
    return &HcString{
    }
}


/**
 * Concatinates strings "s1" and "s2" to a new string "s1s2".
 *
 * This function uses copy() to combine the two strings into a new buffer.
 */
func (hcStr HcString) Concat(s1, s2 string) string {

    var b []byte
    b = make([]byte, len(s1) + len(s2))

    copy(b[0:], s1)
    copy(b[len(s1):], s2)

    return string(b)

}


/**
 * Concatinates strings "s1" and "s2" to a new string "s1s2".
 *
 * This function uses sprintf() to generate a new string.
 */
func (hcStr HcString) Glue(s1, s2 string) string {

    return fmt.Sprintf("%s%s", s1, s2)

}
