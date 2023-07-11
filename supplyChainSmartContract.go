package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// SCSC illustrates how to implement our model based on Hyperledger Fabric framework, which is one of the most popular permissioned blockchain frameworks.
type supplyChainContract struct{
}

// When deploying the smart contract, this function is used to initilize the contract configurations
func (con *supplyChainContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if args[0] != "deploy" {
		return shim.Error("Incorrect command! This smart contract is for deploying supply chain contracts.")
	}
	if len(args[1:]) != 6 {
		return shim.Error("Incorrect number of arguments! Expecting six parameters for the contract: quantity_1, payment_1, quantity_2s0, payment_2s0, quantity_2s1, payment_2s1.")
	}

	quantity_1 := args[1]
	payment_1 := args[2]
	quantity_2_s0 := args[3]
	payment_2_s0 := args[4]
	quantity_2_s1 := args[5]
	payment_2_s1 := args[6]
	buyer_balance := strconv.FormatFloat(strconv.ParseFloat(payment_1, 64) + strconv.ParseFloat(payment_2_s1, 64), 'g', 10, 64)
	supplier_balance := "0"

	err_q1 := stub.PutState("quantity_1", []byte(quantity_1))
	if err_q1 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify quantity for the first period: ", quantity_1))
	}

	err_p1 := stub.PutState("payment_1", []byte(payment_1))
	if err_p1 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify payment for the first period: ", payment_1))
	}

	err_q2s0 := stub.PutState("quantity_2s0", []byte(quantity_2_s0))
	if err_q2s0 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify quantity for the second period with a sinal of 0: ", quantity_2_s0))
	}

	err_p2s0 := stub.PutState("payment_2s0", []byte(payment_2_s0))
	if err_p2s0 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify payment for the second period with a sinal of 0: ", payment_2_s0))
	}

	err_q2s1 := stub.PutState("quantity_2s1", []byte(quantity_2_s1))
	if err_q2s1 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify quantity for the second period with a sinal of 1: ", quantity_2_s1))
	}

	err_p2s1 := stub.PutState("payment_2s1", []byte(payment_2_s1))
	if err_p2s1 != nil {
		return shim.Error(fmt.Sprintf("Failed to specify payment for the second period with a sinal of 1: ", payment_2_s1))
	}

	err_buyerb := stub.PutState("buyer_balance", []byte(buyer_balance))
	if err_buyerb != nil {
		return shim.Error(fmt.Sprintf("Failed to specify account balance for the buyer: ", buyer_balance))
	}

	err_supplierb := stub.PutState("supplier_balance", []byte(supplier_balance))
	if err_supplierb != nil {
		return shim.Error(fmt.Sprintf("Failed to specify account balance for the supplier: ", supplier_balance))
	}

	return shim.Success(nil)
}

func (con *supplyChainContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "sendGoods" {
		result, err = sendGoods(stub, args)
	} else if fn == "signal" {
		result, err = signal(stub, args)
	} else if fn == "realize" {
		result, err = realize(stub)
	} else if fn == "query" {
		result, err = query(stub, args)
	} else {
		errMsg := "Unsupported functions: " + fn
		return shim.Error(errMsg)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))
}

func sendGoods(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments for sending goods. Expecting 1 parameter.")
	}

	predefinedValue, err := stub.GetState("quantity_1")
	if err != nil {
		return "", fmt.Errorf("Failed to query first period quantity with error: %s", err)
	}

	buyer_balance_1, err := stub.GetState("buyer_balance")
	if err != nil {
		return "", fmt.Errorf("Failed to query buyer's account balance with error: %s", err)
	}

	if args[0] == predefinedValue {
		err_update_supplierb := stub.PutState("supplier_balance", []byte(predefinedValue))
		if err_update_supplierb != nil {
			return shim.Error(fmt.Sprintf("Failed to update the account balance for the supplier: ", predefinedValue))
		}

		buyer_balance_2 := strconv.FormatFloat(strconv.ParseFloat(buyer_balance_1, 64) - strconv.ParseFloat(predefinedValue, 64), 'g', 10, 64)
		err_update_buyerb := stub.PutState("buyer_balance", []byte(buyer_balance_2))
		if err_update_buyerb != nil {
			return shim.Error(fmt.Sprintf("Failed to update the account balance for the buyer: ", buyer_balance_2))
		}

		return "Sending goods success", nil

	} else {
		return "", fmt.Errorf("Wrong quantity of goods to send: %s. Please check you contract.", args[0])
	}
}

func signal(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments for signal. Expecting 1 parameter.")
	}

	if (args[0] == "0" || args[0] == "1") {
		err_signal := stub.PutState("signal", []byte(args[0]))
		if err_signal != nil {
			return shim.Error(fmt.Sprintf("Failed to record the signal: ", args[0]))
		}
	} else {
		err_wrong_signal := "Unsupported signal value: " + args[0]
		return shim.Error(err_wrong_signal)
	}

	return "Signaling success", nil
}

func realize(stub shim.ChaincodeStubInterface) (string, error) {
	signal_value, err := stub.GetState("signal")
	if err != nil {
		return "", fmt.Errorf("Failed to query the signal value.")
	}

	buyer_balance_2, err := stub.GetState("buyer_balance")
	if err != nil {
		return "", fmt.Errorf("Failed to query buyer's account balance with error: %s", err)
	}

	supplier_balance_2, err := stub.GetState("supplier_balance")
	if err != nil {
		return "", fmt.Errorf("Failed to query supplier's account balance with error: %s", err)
	}

	if signal_value == "0" {
		payment_2, err := stub.GetState("payment_2s0")
		if err != nil {
			return "", fmt.Errorf("Failed to query the second payment given a signal of 0: %s", err)
		}
	} else if signal_value == "1" {
		payment_2, err := stub.GetState("payment_2s1")
		if err != nil {
			return "", fmt.Errorf("Failed to query the second payment given a signal of 1: %s", err)
		}
	}

	buyer_balance_final := strconv.FormatFloat(strconv.ParseFloat(buyer_balance_2, 64) - strconv.ParseFloat(payment_2, 64), 'g', 10, 64)
	err_update_buyerb2 := stub.PutState("buyer_balance", []byte(buyer_balance_final))
	if err_update_buyerb2 != nil {
		return shim.Error(fmt.Sprintf("Failed to update the account balance for the buyer: ", buyer_balance_final))
	}

	supplier_balance_final := strconv.FormatFloat(strconv.ParseFloat(supplier_balance_2, 64) + strconv.ParseFloat(payment_2, 64), 'g', 10, 64)
	err_update_supplierb2 := stub.PutState("supplier_balance", []byte(supplier_balance_final))
	if err_update_supplierb2 != nil {
		return shim.Error(fmt.Sprintf("Failed to update the account balance for the supplier: ", supplier_balance_final))
	}

	return "Realize the second payment success", nil
}

func query(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting 1 parameter, which is one of the items pre-defined in your contract.")
	}

	if (args[0] == "quantity_1" || args[0] == "payment_1" || args[0] == "quantity_2s0" || args[0] == "payment_2s0" || args[0] == "quantity_2s1" || args[0] == "payment_2s1" || args[0] == "signal") {
		value, err := stub.GetState(args[0])
		if err != nil {
			return "", fmt.Errorf("Failed to query the item: %s with error: %s", args[0], err)
		}

		return string(value), nil
	} else {
		errQueryMsg := "Unsupported item: " + args[0]
		return shim.Error(errQueryMsg)
	}
}

func main(){
	if err := shim.Start(new(supplyChainContract)); err != nil {
		fmt.Printf("Error starting supply chain contrac chaincode: %s", err)
	}
}