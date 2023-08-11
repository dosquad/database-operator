package controller

import (
	"context"
	"errors"
	"hash/crc32"
	"strings"

	"github.com/dosquad/database-operator/accountsvr"
	v1 "github.com/dosquad/database-operator/api/v1"
	"github.com/oklog/ulid/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ()

// newDatabaseAccountName returns a newly generated database/username.
func newDatabaseAccountName(_ context.Context) v1.PostgreSQLResourceName {
	return v1.PostgreSQLResourceName(strings.ToLower("k8s_" + ulid.Make().String()))
}

func setSecretKV(secret *corev1.Secret, key, value string) {
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[key] = []byte(value)
}

type secretFunc func(secret *corev1.Secret) error

func secretRun(
	ctx context.Context,
	r client.Reader,
	w client.Writer,
	accountSvr *accountsvr.Server,
	dbAccount *v1.DatabaseAccount,
	f secretFunc,
) error {
	logger := log.FromContext(ctx)

	secret, secretErr := secretGet(ctx, r, dbAccount)
	if errors.Is(secretErr, ErrNewSecret) {
		accountSvr.CopyInitConfigToSecret(secret)
		if err := w.Create(ctx, secret); err != nil {
			logger.V(1).Error(err, "unable to create secret")

			return err
		}
	} else if secretErr != nil {
		return secretErr
	}

	preChecksum := crc32.ChecksumIEEE([]byte(secret.String()))

	if err := f(secret); err != nil {
		return err
	}

	postChecksum := crc32.ChecksumIEEE([]byte(secret.String()))

	// logger.V(1).Info("Checksum Comparison", "preChecksum", preChecksum, "postChecksum", postChecksum)

	if preChecksum != postChecksum {
		logger.V(1).Info("Secret updated",
			"generation", secret.Generation,
			"resourceVersion", secret.ResourceVersion,
			"preChecksum", preChecksum, "postChecksum", postChecksum,
		)
		if err := w.Update(ctx, secret); err != nil {
			logger.V(1).Error(err, "unable to update secret")

			return err
		}
	}

	return nil
}

func secretGetByName(ctx context.Context, r client.Reader, name types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	if err := r.Get(ctx, name, secret); err != nil {
		return secret, err
	}

	return secret, nil
}

func secretGet(ctx context.Context, r client.Reader, dbAccount *v1.DatabaseAccount) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)

	secret := &corev1.Secret{}
	if err := r.Get(ctx, dbAccount.GetSecretName(), secret); apierrors.IsNotFound(err) {
		logger.V(1).Info("call:secretGet()",
			"dbAccount.Spec.OnDelete", dbAccount.GetSpecOnDelete(),
		)
		// user := uuid.NewUUID()
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        dbAccount.GetSecretName().Name,
				Namespace:   dbAccount.Namespace,
				Annotations: dbAccount.Spec.SecretTemplate.Annotations,
				Labels:      dbAccount.Spec.SecretTemplate.Labels,
			},
			StringData: map[string]string{},
			Data:       map[string][]byte{},
			Type:       secretType,
		}
		if dbAccount.GetSpecOnDelete() == v1.OnDeleteDelete {
			secretAddOwnerRefs(secret, dbAccount)
		}

		return secret, ErrNewSecret
	} else if err != nil {
		return secret, err
	}

	return secret, nil
}

func secretAddOwnerRefs(secret *corev1.Secret, dbAccount *v1.DatabaseAccount) {
	secret.OwnerReferences = []metav1.OwnerReference{*dbAccount.GetReference()}
}
