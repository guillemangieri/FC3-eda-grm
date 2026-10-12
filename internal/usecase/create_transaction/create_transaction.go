package create_transaction

import (
	"context"
	"fmt"
	"sync"

	"github.com.br/devfullcycle/fc-ms-wallet/internal/entity"
	"github.com.br/devfullcycle/fc-ms-wallet/internal/event"
	"github.com.br/devfullcycle/fc-ms-wallet/internal/gateway"
	"github.com.br/devfullcycle/fc-ms-wallet/pkg/events"
	"github.com.br/devfullcycle/fc-ms-wallet/pkg/uow"
)

type CreateTransactionInputDTO struct {
	AccountIDFrom string  `json:"account_id_from"`
	AccountIDTo   string  `json:"account_id_to"`
	Amount        float64 `json:"amount"`
}

type CreateTransactionOutputDTO struct {
	ID            string  `json:"id"`
	AccountIDFrom string  `json:"account_id_from"`
	AccountIDTo   string  `json:"account_id_to"`
	Amount        float64 `json:"amount"`
}

type BalanceUpdatedOutputDTO struct {
	AccountIDFrom        string  `json:"account_id_from"`
	AccountIDTo          string  `json:"account_id_to"`
	BalanceAccountIDFrom float64 `json:"balance_account_id_from"`
	BalanceAccountIDTo   float64 `json:"balance_account_id_to"`
}

type CreateTransactionUseCase struct {
	Uow             uow.UowInterface
	EventDispatcher events.EventDispatcherInterface
	mu              sync.Mutex
}

func NewCreateTransactionUseCase(
	Uow uow.UowInterface,
	eventDispatcher events.EventDispatcherInterface,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		Uow:             Uow,
		EventDispatcher: eventDispatcher,
	}
}

func (uc *CreateTransactionUseCase) Execute(ctx context.Context, input CreateTransactionInputDTO) (*CreateTransactionOutputDTO, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	output := &CreateTransactionOutputDTO{}
	balanceUpdatedOutput := &BalanceUpdatedOutputDTO{}
	err := uc.Uow.Do(ctx, func(txUow *uow.Uow) error {
		accountRepository, err := getAccountRepository(ctx, txUow)
		if err != nil {
			return err
		}
		transactionRepository, err := getTransactionRepository(ctx, txUow)
		if err != nil {
			return err
		}

		accountFrom, err := accountRepository.FindByID(input.AccountIDFrom)
		if err != nil {
			return err
		}
		accountTo, err := accountRepository.FindByID(input.AccountIDTo)
		if err != nil {
			return err
		}
		transaction, err := entity.NewTransaction(accountFrom, accountTo, input.Amount)
		if err != nil {
			return err
		}

		err = accountRepository.UpdateBalance(accountFrom)
		if err != nil {
			return err
		}

		err = accountRepository.UpdateBalance(accountTo)
		if err != nil {
			return err
		}

		err = transactionRepository.Create(transaction)
		if err != nil {
			return err
		}
		output.ID = transaction.ID
		output.AccountIDFrom = input.AccountIDFrom
		output.AccountIDTo = input.AccountIDTo
		output.Amount = input.Amount

		balanceUpdatedOutput.AccountIDFrom = input.AccountIDFrom
		balanceUpdatedOutput.AccountIDTo = input.AccountIDTo
		balanceUpdatedOutput.BalanceAccountIDFrom = accountFrom.Balance
		balanceUpdatedOutput.BalanceAccountIDTo = accountTo.Balance

		transactionCreated := event.NewTransactionCreated()
		transactionCreated.SetPayload(output)
		if err := uc.EventDispatcher.Dispatch(transactionCreated); err != nil {
			return err
		}

		balanceUpdated := event.NewBalanceUpdated()
		balanceUpdated.SetPayload(balanceUpdatedOutput)
		if err := uc.EventDispatcher.Dispatch(balanceUpdated); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func getAccountRepository(ctx context.Context, txUow *uow.Uow) (gateway.AccountGateway, error) {
	repo, err := txUow.GetRepository(ctx, "AccountDB")
	if err != nil {
		return nil, err
	}
	accountRepository, ok := repo.(gateway.AccountGateway)
	if !ok {
		return nil, fmt.Errorf("AccountDB has unexpected type %T", repo)
	}
	return accountRepository, nil
}

func getTransactionRepository(ctx context.Context, txUow *uow.Uow) (gateway.TransactionGateway, error) {
	repo, err := txUow.GetRepository(ctx, "TransactionDB")
	if err != nil {
		return nil, err
	}
	transactionRepository, ok := repo.(gateway.TransactionGateway)
	if !ok {
		return nil, fmt.Errorf("TransactionDB has unexpected type %T", repo)
	}
	return transactionRepository, nil
}
