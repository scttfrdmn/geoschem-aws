package ssh

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type KeyPairManager struct {
	ec2Client *ec2.Client
}

// NewKeyPairManager creates a new key pair manager
func NewKeyPairManager(ec2Client *ec2.Client) *KeyPairManager {
	return &KeyPairManager{
		ec2Client: ec2Client,
	}
}

// CreateKeyPair creates a new key pair in AWS and returns the private key
func (kpm *KeyPairManager) CreateKeyPair(ctx context.Context, keyName string) (*KeyPair, error) {
	// Generate local key pair first
	keyPair, err := GenerateKeyPair(keyName)
	if err != nil {
		return nil, fmt.Errorf("generating key pair: %w", err)
	}

	// Import public key to AWS
	input := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: []byte(keyPair.PublicKey),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeKeyPair,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(keyName)},
					{Key: aws.String("Project"), Value: aws.String("geoschem-aws")},
					{Key: aws.String("Purpose"), Value: aws.String("builder-ssh")},
				},
			},
		},
	}

	_, err = kpm.ec2Client.ImportKeyPair(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("importing key pair to AWS: %w", err)
	}

	return keyPair, nil
}

// KeyPairExists checks if a key pair exists in AWS
func (kpm *KeyPairManager) KeyPairExists(ctx context.Context, keyName string) (bool, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{keyName},
	}

	_, err := kpm.ec2Client.DescribeKeyPairs(ctx, input)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("checking key pair existence: %w", err)
	}

	return true, nil
}

// DeleteKeyPair deletes a key pair from AWS
func (kpm *KeyPairManager) DeleteKeyPair(ctx context.Context, keyName string) error {
	input := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(keyName),
	}

	_, err := kpm.ec2Client.DeleteKeyPair(ctx, input)
	if err != nil {
		return fmt.Errorf("deleting key pair: %w", err)
	}

	return nil
}

// ListKeyPairs lists all key pairs with the project tag
func (kpm *KeyPairManager) ListKeyPairs(ctx context.Context) ([]string, error) {
	input := &ec2.DescribeKeyPairsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("tag:Project"),
				Values: []string{"geoschem-aws"},
			},
		},
	}

	result, err := kpm.ec2Client.DescribeKeyPairs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("listing key pairs: %w", err)
	}

	var keyNames []string
	for _, keyPair := range result.KeyPairs {
		if keyPair.KeyName != nil {
			keyNames = append(keyNames, *keyPair.KeyName)
		}
	}

	return keyNames, nil
}

// GetOrCreateKeyPair gets an existing key pair or creates a new one
func (kpm *KeyPairManager) GetOrCreateKeyPair(ctx context.Context, keyName, privateKeyPath string) error {
	// Check if key pair exists in AWS
	exists, err := kpm.KeyPairExists(ctx, keyName)
	if err != nil {
		return fmt.Errorf("checking key pair existence: %w", err)
	}

	// Check if private key file exists locally
	if exists {
		// Verify local private key file exists
		if _, err := os.Stat(privateKeyPath); err == nil {
			// Both AWS key pair and local private key exist
			return nil
		}
		// AWS key pair exists but no local private key - this is a problem
		return fmt.Errorf("key pair %s exists in AWS but no local private key found at %s", keyName, privateKeyPath)
	}

	// Neither exists, create new key pair
	keyPair, err := kpm.CreateKeyPair(ctx, keyName)
	if err != nil {
		return fmt.Errorf("creating key pair: %w", err)
	}

	// Save to local files
	err = SaveKeyPairToFile(keyPair, privateKeyPath)
	if err != nil {
		// Clean up AWS key pair if local save fails
		kpm.DeleteKeyPair(ctx, keyName)
		return fmt.Errorf("saving key pair to file: %w", err)
	}

	fmt.Printf("Created new key pair %s and saved to %s\n", keyName, privateKeyPath)
	return nil
}