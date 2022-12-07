package service

import (
	"bufio"
	"context"
	"fmt"
	Client "github.com/tonx22/lancktask/client"
	"os"
	"strings"
)

func RunAsClient(search, methodType, token *string) error {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	client, err := Client.NewGRPCClient(*token)
	if err != nil {
		return fmt.Errorf("Failed to create grpc client: %v", err)
	}

	s := strings.ReplaceAll(*search, " ", "")
	if len(s) == 0 {
		return fmt.Errorf("No phone numbers to search")
	}
	numbers := strings.Split(s, ",")

	if *methodType == "streaming" {
		res, err := client.StreamingGetCodeByNumber(context.TODO(), numbers)
		if err != nil {
			return fmt.Errorf("failed to get codes by numbers: %v", err)
		}
		for _, code := range *res {
			fmt.Fprintln(out, code)
		}

	} else {
		for _, number := range numbers {
			res, err := client.GetCodeByNumber(context.TODO(), number)
			if err != nil {
				return fmt.Errorf("failed to get code by number: %v", err)
			}
			fmt.Fprintln(out, *res)
		}
	}
	return nil
}
