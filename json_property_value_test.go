package cfbackup_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotalservices/cfbackup"
)

var testMessage = `
[{
	"value": "this is a string"
}, {
	"value": 12345
}, {
	"value": {
		"map": "map_value"
	}
}, {
	"value": [12345, "string_value", {
		"map": "map_value"
	}]
}]
`

type MessageValue struct {
	Value PropertyValue `json:"value"`
}
type Message []MessageValue

var _ = Describe("Given a composite value message", func() {
	Context("When json unmarshal this message", func() {
		var message Message
		err := json.Unmarshal([]byte(testMessage), &message)
		It("should not get an error", func() {
			Ω(err).ShouldNot(HaveOccurred())
		})
		It("should get a string", func() {
			Ω(message[0].Value.StringValue).Should(Equal("this is a string"))
		})
		It("should get a number", func() {
			Ω(message[1].Value.IntValue).Should(Equal(uint64(12345)))
		})
		It("should get a map with correct map", func() {
			Ω(len(message[2].Value.MapValue)).Should(Equal(1))
		})
		It("should get an array with correct values", func() {
			Ω(len(message[3].Value.ArrayValue)).Should(Equal(3))
		})
	})
})
