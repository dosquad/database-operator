package controller

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"text/template"

	"github.com/dosquad/database-operator/accountsvr"
	dbov1 "github.com/dosquad/database-operator/api/v1"
	"github.com/oklog/ulid/v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	labelNamePartOf           = "app.kubernetes.io/part-of"
	defaultPartOf             = "database-operator-relay"
	defaultPostgresqlPort     = 5432
	defaultPostgresqlPortName = "postgresql"
	databaseOwnerKey          = "dba-name"
)

// NewDatabaseAccountName returns a newly generated database/username.
func NewDatabaseAccountName(_ context.Context) dbov1.PostgreSQLResourceName {
	return dbov1.PostgreSQLResourceName(strings.ToLower("k8s_" + ulid.Make().String()))
}

func SetSecretKV(secret *corev1.Secret, key, value string) {
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[key] = []byte(value)
}

func GetSecretKV(secret *corev1.Secret, key string) string {
	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	if secret.StringData == nil {
		secret.StringData = make(map[string]string)
	}

	if v, ok := secret.Data[key]; ok {
		return string(v)
	}

	return ""
}

const (
	pgBouncerConfTemplate = `[databases]
* = host={{.Host}} port=5432 user={{.User}} password={{.Password}}

[pgbouncer]
listen_addr = 0.0.0.0
listen_port = {{.Port}}
unix_socket_dir =
user = postgres
auth_file = /etc/pgbouncer/userlist.txt
auth_type = md5
ignore_startup_parameters = extra_float_digits

# Log settings
admin_users = {{.User}}
stats_users = {{.User}}
`
)

func AddPGBouncerConf(
	accountSvr accountsvr.Server,
	secret *corev1.Secret,
) {
	tmpl, tmplErr := template.New("").Parse(pgBouncerConfTemplate)
	if tmplErr != nil {
		return
	}

	data := struct {
		Host, Port, User, Password string
	}{
		Host:     accountSvr.GetDatabaseHostConfig(),
		Port:     strconv.Itoa(defaultPostgresqlPort),
		User:     GetSecretKV(secret, accountsvr.DatabaseKeyUsername),
		Password: GetSecretKV(secret, accountsvr.DatabaseKeyPassword),
	}

	sb := &strings.Builder{}
	if err := tmpl.Execute(sb, data); err != nil {
		return
	}

	SetSecretKV(secret, accountsvr.DatabaseKeyPGBouncerConf, sb.String())

	SetSecretKV(
		secret, accountsvr.DatabaseKeyPGBouncerUsers,
		fmt.Sprintf("\"%s\" \"%s\"\n",
			GetSecretKV(secret, accountsvr.DatabaseKeyUsername),
			GetSecretKV(secret, accountsvr.DatabaseKeyPassword),
		),
	)
}

type SecretFunc func(secret *corev1.Secret) error

func SecretRun(
	ctx context.Context,
	r client.Reader,
	w client.Writer,
	accountSvr accountsvr.Server,
	dbAccount *dbov1.DatabaseAccount,
	f SecretFunc,
) error {
	logger := log.FromContext(ctx)

	secret, secretErr := SecretGet(ctx, r, dbAccount)
	if errors.Is(secretErr, ErrNewSecret) {
		accountSvr.CopyInitConfigToSecret(dbAccount, secret)
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

func SecretGetByName(ctx context.Context, r client.Reader, name types.NamespacedName) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	if err := r.Get(ctx, name, secret); err != nil {
		return secret, err
	}

	return secret, nil
}

func SecretGet(ctx context.Context, r client.Reader, dbAccount *dbov1.DatabaseAccount) (*corev1.Secret, error) {
	logger := log.FromContext(ctx)

	secret := &corev1.Secret{}
	if err := r.Get(ctx, dbAccount.GetSecretName(), secret); apierrors.IsNotFound(err) {
		logger.V(1).Info("call:SecretGet()",
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
		if dbAccount.GetSpecOnDelete() == dbov1.OnDeleteDelete {
			SecretAddOwnerRefs(secret, dbAccount)
		}

		return secret, ErrNewSecret
	} else if err != nil {
		return secret, err
	}

	return secret, nil
}

func SecretAddOwnerRefs(secret *corev1.Secret, dbAccount *dbov1.DatabaseAccount) {
	secret.OwnerReferences = []metav1.OwnerReference{*dbAccount.GetReference()}
}

func StatefulSetGet(
	ctx context.Context,
	r client.Reader,
	ctrlConfig *dbov1.DatabaseAccountControllerConfig,
	dbAccount *dbov1.DatabaseAccount,
) (*appsv1.StatefulSet, error) {
	logger := log.FromContext(ctx)

	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, dbAccount.GetStatefulSetName(), statefulSet); apierrors.IsNotFound(err) {
		logger.V(1).Info("call:StatefulSetGet()")

		podSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: ctrlConfig.GetRelayImage(),
					Name:  dbAccount.GetStatefulSetName().Name,
					Ports: []corev1.ContainerPort{
						{
							Name:          defaultPostgresqlPortName,
							ContainerPort: defaultPostgresqlPort,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					ImagePullPolicy: "IfNotPresent",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      dbAccount.GetSecretName().Name,
							MountPath: "/etc/pgbouncer/",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: dbAccount.GetSecretName().Name,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: dbAccount.GetSecretName().Name,
						},
					},
				},
			},
		}

		l := labels.Merge(dbAccount.Spec.SecretTemplate.Labels, map[string]string{
			labelNamePartOf:  defaultPartOf,
			databaseOwnerKey: dbAccount.GetStatefulSetName().Name,
		})

		statefulSet = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:            dbAccount.GetStatefulSetName().Name,
				Namespace:       dbAccount.Namespace,
				Annotations:     dbAccount.Spec.SecretTemplate.Annotations,
				Labels:          l,
				OwnerReferences: []metav1.OwnerReference{*dbAccount.GetReference()},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
				Selector: &metav1.LabelSelector{
					MatchLabels: l,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:        dbAccount.GetStatefulSetName().Name,
						Namespace:   dbAccount.Namespace,
						Annotations: dbAccount.Spec.SecretTemplate.Annotations,
						Labels:      l,
					},
					Spec: podSpec,
				},
			},
		}

		return statefulSet, ErrNewStatefulSet
	} else if err != nil {
		return statefulSet, err
	}

	return statefulSet, nil
}

// func ServiceGetByName(ctx context.Context, r client.Reader, name types.NamespacedName) (*corev1.Service, error) {
// 	service := &corev1.Service{}

// 	if err := r.Get(ctx, name, service); err != nil {
// 		return service, err
// 	}

// 	return service, nil
// }

func ServiceGet(
	ctx context.Context,
	r client.Reader,
	dbAccount *dbov1.DatabaseAccount,
) (*corev1.Service, error) {
	logger := log.FromContext(ctx)

	service := &corev1.Service{}
	if err := r.Get(ctx, dbAccount.GetStatefulSetName(), service); apierrors.IsNotFound(err) {
		logger.V(1).Info("call:ServiceGet()")

		l := labels.Merge(dbAccount.Spec.SecretTemplate.Labels, map[string]string{
			labelNamePartOf:  defaultPartOf,
			databaseOwnerKey: dbAccount.GetStatefulSetName().Name,
		})

		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:            dbAccount.GetStatefulSetName().Name,
				Namespace:       dbAccount.Namespace,
				Annotations:     dbAccount.Spec.SecretTemplate.Annotations,
				Labels:          l,
				OwnerReferences: []metav1.OwnerReference{*dbAccount.GetReference()},
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       defaultPostgresqlPortName,
						Port:       defaultPostgresqlPort,
						TargetPort: intstr.FromString(defaultPostgresqlPortName),
						Protocol:   corev1.ProtocolTCP,
					},
				},
				Selector: map[string]string{
					databaseOwnerKey: dbAccount.GetStatefulSetName().Name,
				},
			},
		}

		return service, ErrNewService
	} else if err != nil {
		return service, err
	}

	return service, nil
}

// func stringOnly(in string, _ bool) string {
// 	return in
// }

// func getLabel(labels map[string]string, key string) (string, bool) {
// 	if v, ok := labels[key]; ok {
// 		return v, ok
// 	}

// 	return "", false
// }
