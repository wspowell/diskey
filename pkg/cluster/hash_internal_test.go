package cluster

// TODO
// func Test_getClosestAddress(t *testing.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		description     string
// 		key             string
// 		addresses       []Address
// 		expectedAddress Address
// 	}{
// 		{
// 			description: "one address",
// 			key:         "key",
// 			addresses: []Address{
// 				{
// 					Host: "localhost",
// 					Port: "7000",
// 				},
// 			},
// 			expectedAddress: Address{
// 				Host: "localhost",
// 				Port: "7000",
// 			},
// 		},
// 		{
// 			description: "two addresses",
// 			key:         "key",
// 			addresses: []Address{
// 				{
// 					Host: "localhost",
// 					Port: "7000",
// 				},
// 				{
// 					Host: "localhost",
// 					Port: "7001",
// 				},
// 			},
// 			expectedAddress: Address{
// 				Host: "localhost",
// 				Port: "7001",
// 			},
// 		},
// 	}
// 	for index := range testCases {
// 		testCase := testCases[index]
// 		t.Run(testCase.description, func(t *testing.T) {
// 			t.Parallel()

// 			actualAddress := getClosestAddress(testCase.key, testCase.addresses)
// 			assert.Equal(t, testCase.expectedAddress, actualAddress)
// 		})
// 	}
// }
