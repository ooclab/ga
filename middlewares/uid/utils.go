package uid

import (
	"github.com/Sirupsen/logrus"
	"github.com/ooclab/ga/service/etcd"
)

func loadPublicKeyFromEtcd(publicKeyPath string) ([]byte, error) {
	// get public key
	session, err := etcd.GetSession()
	if err != nil {
		logrus.Errorf("get etcd session failed: %s\n", err)
		return nil, err
	}

	pubKey, err := session.Get(publicKeyPath)
	if err != nil {
		logrus.Errorf("get public key from etcd failed: %s\n", err)
		return nil, err
	}
	logrus.Debugf("load public key (%s) success\n", publicKeyPath)

	return []byte(pubKey), nil
}
