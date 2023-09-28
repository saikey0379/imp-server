package route

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/saikey0379/go-json-rest/rest"
	"golang.org/x/net/context"

	"github.com/saikey0379/imp-server/pkg/known"
	"github.com/saikey0379/imp-server/pkg/middleware"
	"github.com/saikey0379/imp-server/pkg/model"
	"github.com/saikey0379/imp-server/pkg/utils"
)

type ClusterId struct {
	ID int `json:"ID"`
}

type CertListPageReq struct {
	ID          int         `json:"id"`
	ClusterId   int         `json:"ClusterId"`
	ClusterIds  []ClusterId `json:"ClusterIds"`
	DomainId    int         `json:"DomainId"`
	DomainName  string      `json:"DomainName"`
	RouteId     int         `json:"RouteId"`
	AccessToken string      `json:"AccessToken"`
	Keyword     string      `json:"keyword"`
	AccessType  string      `json:"AccessType"`
	Limit       uint
	Offset      uint
}

func getCertConditions(req CertListPageReq) string {
	var where []string
	if req.ID > 0 {
		where = append(where, fmt.Sprintf("proxy_cert.id = %d", req.ID))
	}
	if req.ClusterId > 0 {
		where = append(where, fmt.Sprintf("proxy_cert.cluster_ids like %s%d%s", "'%", req.ClusterId, "%'"))
	}
	for _, v := range req.ClusterIds {
		if v.ID > 0 {
			where = append(where, fmt.Sprintf("proxy_cert.cluster_ids like %s%d%s", "'%", v.ID, "%'"))
		}
	}
	if req.DomainName != "" {
		var name string
		for k, i := range strings.Split(req.DomainName, ".") {
			if k > 0 {
				name += "." + strings.TrimSpace(i)
			}
		}
		where = append(where, fmt.Sprintf("proxy_cert.name like %s", "'%"+name+"%'"))
	}
	if req.Keyword = strings.TrimSpace(req.Keyword); req.Keyword != "" {
		where = append(where, fmt.Sprintf("( proxy_cert.name like %s or proxy_cert.description like %s )", "'%"+req.Keyword+"%'", "'%"+req.Keyword+"%'"))
	}
	if len(where) > 0 {
		return " where " + strings.Join(where, " and ")
	} else {
		return ""
	}
}

func GetCertSelectByDomainNameAndClusterIds(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info CertListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mods, err := repo.GetCertListWithPage(0, 0, getCertConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}

	type Cert struct {
		ID   int
		Name string
	}

	var cert Cert
	var certs []Cert
	for _, i := range mods {
		compileRegex := regexp.MustCompile(strings.Replace(i.Name, "*", ".*", -1))
		matchArr := compileRegex.FindStringSubmatch(info.DomainName)
		if len(matchArr) > 0 {
			cert.ID = i.ID
			cert.Name = i.Name
			certs = append(certs, cert)
		}
	}
	result := make(map[string]interface{})
	result["list"] = certs
	result["ClusterId"] = info.ClusterId
	//总条数
	count := len(certs)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetCertList(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info CertListPageReq
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	type ClusterId struct {
		ID int
	}
	type Cert struct {
		ID          int
		Name        string
		ClusterIds  []ClusterId
		Description string
		Manager     string
		Status      string
		NotBefore   utils.ISOTime
		NotAfter    utils.ISOTime
		UpdatedAt   utils.ISOTime
	}
	mods, err := repo.GetCertListWithPage(info.Limit, info.Offset, getCertConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var cert Cert
	var certs []Cert
	for _, i := range mods {
		cert.ID = i.ID
		cert.Name = i.Name
		cert.Description = i.Description
		cert.Manager = i.Manager
		cert.NotBefore = utils.ISOTime(i.NotBefore)
		cert.NotAfter = utils.ISOTime(i.NotAfter)
		cert.UpdatedAt = utils.ISOTime(i.UpdatedAt)
		var clusterid ClusterId
		var clusterids []ClusterId
		for _, i := range strings.Split(i.ClusterIds, ",") {
			clusterid.ID, err = strconv.Atoi(i)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			clusterids = append(clusterids, clusterid)
		}
		cert.ClusterIds = clusterids
		if timeSub(i.NotAfter, time.Now()) > 30 {
			cert.Status = "有效"
			certs = append(certs, cert)
		} else if timeSub(i.NotAfter, time.Now()) > 0 {
			cert.Status = "即将过期"
			var tmp []Cert
			tmp = append(tmp, cert)
			tmp = append(tmp, certs...)
			certs = tmp
		} else {
			cert.Status = "已过期"
			var tmp []Cert
			tmp = append(tmp, cert)
			tmp = append(tmp, certs...)
			certs = tmp
		}
	}
	result := make(map[string]interface{})
	result["list"] = certs
	result["ClusterId"] = info.ClusterId
	//总条数
	count, err := repo.CountCert(getCertConditions(info))
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	result["recordCount"] = count
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": result})
}

func GetCertById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          int
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}

	mod, err := repo.GetCertById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	type ClusterId struct {
		ID int
	}
	type Cert struct {
		ID          int
		Name        string
		Manager     string
		Description string
		ClusterIds  []ClusterId
		FileCert    string
		ContentCert []string
		FileKey     string
		ContentKey  []string
		NotBefore   utils.ISOTime
		NotAfter    utils.ISOTime
		CreatedAt   utils.ISOTime
		UpdatedAt   utils.ISOTime
		Domains     []model.DomainUs
	}
	var cert Cert
	cert.ID = info.ID
	cert.Name = mod.Name
	cert.Description = mod.Description
	cert.FileCert = mod.FileCert
	cert.FileKey = mod.FileKey
	cert.NotBefore = utils.ISOTime(mod.NotBefore)
	cert.NotAfter = utils.ISOTime(mod.NotAfter)
	cert.Manager = mod.Manager
	cert.CreatedAt = utils.ISOTime(mod.CreatedAt)
	cert.UpdatedAt = utils.ISOTime(mod.UpdatedAt)
	cert.ContentCert = append(cert.ContentCert, mod.ContentCert)
	cert.ContentKey = append(cert.ContentKey, mod.ContentKey)
	var clusterid ClusterId
	var clusterids []ClusterId
	for _, i := range strings.Split(mod.ClusterIds, ",") {
		clusterid.ID, err = strconv.Atoi(i)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		clusterids = append(clusterids, clusterid)
	}
	cert.ClusterIds = clusterids
	domainUs, err := repo.GetDomainListByCertId(info.ID)
	if len(domainUs) > 0 && err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	cert.Domains = domainUs
	w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功", "Content": cert})
}

// 添加
func AddCert(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type ClusterId struct {
		ID int
	}
	type Cert struct {
		Description string
		ClusterIds  []ClusterId
		FileCert    string
		ContentCert []string
		FileKey     string
		ContentKey  []string
		Manager     string
		AccessToken string
	}
	var info Cert
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}
	var mod model.Cert
	mod.ID = int(time.Now().Unix())
	mod.Description = strings.TrimSpace(info.Description)
	mod.FileCert = strings.TrimSpace(info.FileCert)
	mod.ContentCert = info.ContentCert[0]
	mod.FileKey = strings.TrimSpace(info.FileKey)
	mod.ContentKey = info.ContentKey[0]
	mod.Manager = info.Manager
	//证书解析
	certBlock, _ := pem.Decode([]byte(mod.ContentCert))
	if certBlock == nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书格式异常！请确认"})
		return
	}
	certBody, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Parse certificate"})
		return
	}
	var names string
	for i, v := range certBody.DNSNames {
		if i == 0 {
			names = v
		} else {
			names = names + "," + v
		}
	}
	mod.Name = names
	mod.NotBefore = certBody.NotBefore
	mod.NotAfter = certBody.NotAfter
	//根据证书获取公钥
	pubkey_cert, err := x509.MarshalPKIXPublicKey(certBody.PublicKey)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Parse public key from certificate"})
		return
	}
	pubKeyBlk_cert := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubkey_cert,
	}
	pubkeypem_cert := string(pem.EncodeToMemory(&pubKeyBlk_cert))
	//根据私钥获取公钥
	keyBlock, _ := pem.Decode([]byte(mod.ContentKey))
	if keyBlock == nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥格式异常！请确认"})
		return
	}

	// 解析RSA私钥
	privateKeyPKCS8, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		keyBody, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析异常"})
			return
		}
		pubkey := &keyBody.PublicKey
		pubkey_key, err := x509.MarshalPKIXPublicKey(pubkey)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析公钥异常"})
			return
		}
		pubKeyBlk_key := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkey_key,
		}
		pubkeypem_key := string(pem.EncodeToMemory(&pubKeyBlk_key))
		//校验证书和私钥
		if pubkeypem_cert != pubkeypem_key {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书私钥不匹配"})
			return
		}
	} else {
		privateKeyRSA, ok := privateKeyPKCS8.(*rsa.PrivateKey)
		if !ok {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Private key is not RSA private key"})
			return
		}

		pubkey := &privateKeyRSA.PublicKey
		pubkey_key, err := x509.MarshalPKIXPublicKey(pubkey)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析公钥异常"})
			return
		}
		pubKeyBlk_key := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkey_key,
		}
		pubkeypem_key := string(pem.EncodeToMemory(&pubKeyBlk_key))
		//校验证书和私钥
		if pubkeypem_cert != pubkeypem_key {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书私钥不匹配"})
			return
		}
	}

	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(v.ID)
		if i == 0 {
			mod.ClusterIds = clusterid
		} else {
			mod.ClusterIds = mod.ClusterIds + "," + clusterid
		}
		rootDir := known.RootProxy
		var dirssl = path.Join(rootDir, "ssl/"+clusterid)
		err = ioutil.WriteFile(path.Join(dirssl, mod.FileCert), []byte(mod.ContentCert), 0666) //写入文件(字节数组)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
		err = ioutil.WriteFile(path.Join(dirssl, mod.FileKey), []byte(mod.ContentKey), 0666) //写入文件(字节数组)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}
	modcert, errAdd := repo.AddCert(mod)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}
	var message string
	var content string
	contentcert, err := json.Marshal(modcert)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	content = "[add Cert:" + string(contentcert) + "]"

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "add"
	journal.Resource = "cert"
	journal.Content = content
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func DeleteCertById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	var info struct {
		ID          int
		AccessToken string
	}
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误" + err.Error()})
		return
	}
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}
	count, err := repo.CountDomainByCertId(info.ID)
	if count > 0 {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该证书已被调用,请删除相关域名配置"})
		return
	} else if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	mod, err := repo.GetCertById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
	}

	_, err = repo.DeleteCertById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var message string
	contentcert, err := json.Marshal(mod)
	if err != nil {
		message = fmt.Sprintf("[Umarshal failed:%s]", err)
	}
	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "delete"
	journal.Resource = "cert"
	journal.Content = "[delete Cert:" + string(contentcert) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func UpdateCertById(ctx context.Context, w rest.ResponseWriter, r *rest.Request) {
	repo, ok := middleware.RepoFromContext(ctx)
	if !ok {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "内部服务器错误"})
		return
	}
	type ClusterId struct {
		ID int
	}
	type Cert struct {
		ID          int
		Description string
		ClusterIds  []ClusterId
		FileCert    string
		ContentCert []string
		FileKey     string
		ContentKey  []string
		Manager     string
		AccessToken string
	}
	var info Cert
	if err := r.DecodeJSONPayload(&info); err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "参数错误"})
		return
	}
	info.AccessToken = strings.TrimSpace(info.AccessToken)
	user, errVerify := VerifyAccessPurview(info.AccessToken, ctx, true, w, r)
	if errVerify != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errVerify.Error()})
		return
	}
	mod, err := repo.GetCertById(info.ID)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
		return
	}
	var clusterids string
	for i, v := range info.ClusterIds {
		clusterid := strconv.Itoa(int(v.ID))
		if i == 0 {
			clusterids = clusterid
		} else {
			clusterids = clusterids + "," + clusterid
		}
	}

	var mod_update model.Cert
	mod_update.Description = strings.TrimSpace(info.Description)
	mod_update.ClusterIds = clusterids
	mod_update.FileCert = strings.TrimSpace(info.FileCert)
	mod_update.ContentCert = info.ContentCert[0]
	mod_update.FileKey = strings.TrimSpace(info.FileKey)
	mod_update.ContentKey = info.ContentKey[0]
	mod_update.Manager = info.Manager

	var updatebool = false
	var message string
	var content string

	if mod_update.Description != mod.Description {
		updatebool = true
		content = content + "[update Description:\"" + mod.Description + "\" to \"" + mod_update.Description + "\"]"
	}
	if clusterids != mod.ClusterIds {
		updatebool = true
		content = content + "[update ClusterIds:\"" + mod.ClusterIds + "\" to \"" + clusterids + "\"]"
	}
	if mod_update.FileCert != mod.FileCert {
		updatebool = true
		content = content + "[update FileCert:\"" + mod.FileCert + "\" to \"" + mod_update.FileCert + "\"]"
	}
	if mod_update.ContentCert != mod.ContentCert {
		updatebool = true
		content = content + "[update ContentCert:\"" + mod.ContentCert + "\" to \"" + mod_update.ContentCert + "\"]"
	}
	if mod_update.FileKey != mod.FileKey {
		updatebool = true
		content = content + "[update FileKey:\"" + mod.FileKey + "\" to \"" + mod_update.FileKey + "\"]"
	}
	if mod_update.ContentKey != mod.ContentKey {
		updatebool = true
		content = content + "[update ContentKey:\"" + mod.ContentKey + "\" to \"" + mod_update.ContentKey + "\"]"
	}
	if mod_update.Manager != mod.Manager {
		updatebool = true
		content = content + "[update Manager:\"" + mod.Manager + "\" to \"" + mod_update.Manager + "\"]"
	}

	if !updatebool {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "数据未更改"})
		return
	}

	if mod_update.FileCert == mod.FileCert && mod_update.ContentCert != mod.ContentCert {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书[" + mod.FileCert + "]内容已修改！请重命名"})
		return
	}
	if mod_update.FileKey == mod.FileKey && mod_update.ContentKey != mod.ContentKey {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "密钥[" + mod.FileKey + "]内容已修改！请重命名"})
		return
	}

	certBlock, _ := pem.Decode([]byte(mod_update.ContentCert))
	if certBlock == nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书格式异常！请确认"})
		return
	}
	//可从剩余判断是否有证书链等，继续解析
	//证书解析
	certBody, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Parse certificate"})
		return
	}
	var names string
	for i, v := range certBody.DNSNames {
		if i == 0 {
			names = v
		} else {
			names = names + "," + v
		}
	}
	//可以根据证书结构解析
	mod_update.Name = names
	if mod_update.Name != mod.Name {
		content = "[update Name:\"" + mod.Name + "\" to \"" + mod_update.Name + "\"]" + content
	}
	mod_update.NotBefore = certBody.NotBefore
	if mod_update.NotBefore.UTC() != mod.NotBefore.UTC() {
		content = content + "[update NotBefore:\"" + mod.NotBefore.UTC().Format("2006-01-02 15:04:05") + "\" to \"" + mod_update.NotBefore.UTC().Format("2006-01-02 15:04:05") + "\"]"
	}
	mod_update.NotAfter = certBody.NotAfter
	if mod_update.NotAfter.UTC() != mod.NotAfter.UTC() {
		content = content + "[update NotAfter:\"" + mod.NotAfter.UTC().Format("2006-01-02 15:04:05") + "\" to \"" + mod_update.NotAfter.UTC().Format("2006-01-02 15:04:05") + "\"]"
	}
	//根据证书获取公钥
	pubkey_cert, err := x509.MarshalPKIXPublicKey(certBody.PublicKey)
	if err != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Parse public key from certificate"})
		return
	}
	pubKeyBlk_cert := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubkey_cert,
	}
	pubkeypem_cert := string(pem.EncodeToMemory(&pubKeyBlk_cert))
	//根据私钥获取公钥
	keyBlock, _ := pem.Decode([]byte(mod_update.ContentKey))
	if keyBlock == nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥格式异常！请确认"})
		return
	}

	// 解析RSA私钥
	privateKeyPKCS8, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		keyBody, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析异常"})
			return
		}
		pubkey := &keyBody.PublicKey
		pubkey_key, err := x509.MarshalPKIXPublicKey(pubkey)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析公钥异常"})
			return
		}
		pubKeyBlk_key := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkey_key,
		}
		pubkeypem_key := string(pem.EncodeToMemory(&pubKeyBlk_key))
		//校验证书和私钥
		if pubkeypem_cert != pubkeypem_key {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书私钥不匹配"})
			return
		}
	} else {
		privateKeyRSA, ok := privateKeyPKCS8.(*rsa.PrivateKey)
		if !ok {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "Private key is not RSA private key"})
			return
		}

		pubkey := &privateKeyRSA.PublicKey
		pubkey_key, err := x509.MarshalPKIXPublicKey(pubkey)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "私钥解析公钥异常"})
			return
		}
		pubKeyBlk_key := pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkey_key,
		}
		pubkeypem_key := string(pem.EncodeToMemory(&pubKeyBlk_key))
		//校验证书和私钥
		if pubkeypem_cert != pubkeypem_key {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "证书私钥不匹配"})
			return
		}
	}

	dirSSL := path.Join(known.RootProxy, "ssl/")
	for _, v := range strings.Split(mod_update.ClusterIds, ",") {
		var dirSSLCluster = path.Join(dirSSL, v)
		if !strings.Contains(mod.ClusterIds, v) || mod_update.FileCert != mod.FileCert {
			err = utils.WriteFile(path.Join(dirSSLCluster, mod_update.FileCert), mod_update.ContentCert, 0666) //写入文件(字节数组)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		if !strings.Contains(mod.ClusterIds, v) || mod_update.FileKey != mod.FileKey {
			err = utils.WriteFile(path.Join(dirSSLCluster, mod_update.FileKey), mod_update.ContentKey, 0666) //写入文件(字节数组)
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		if mod_update.FileKey != mod.FileKey {
			err = os.Remove(path.Join(dirSSLCluster, mod.FileCert))
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
		if mod_update.FileKey != mod.FileKey {
			err = os.Remove(path.Join(dirSSLCluster, mod.FileKey))
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}

	for _, v := range strings.Split(mod.ClusterIds, ",") {
		if !strings.Contains(clusterids, v) {
			var dirSSLCluster = path.Join(dirSSL, v)
			count, err := repo.CountDomainByCertIdAndClusterId(info.ID, v)
			if count > 0 {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": "该证书已被集群[" + v + "]调用,请删除相关域名证书配置"})
				return
			}
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			err = os.Remove(path.Join(dirSSLCluster, mod.FileCert))
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
			err = os.Remove(path.Join(dirSSLCluster, mod.FileKey))
			if err != nil {
				w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
				return
			}
		}
	}
	_, errAdd := repo.UpdateCertById(info.ID, mod_update.Name, mod_update.Description, mod_update.ClusterIds, mod_update.FileCert, mod_update.ContentCert, mod_update.FileKey, mod_update.ContentKey, mod_update.NotBefore, mod_update.NotAfter, mod_update.Manager)
	if errAdd != nil {
		w.WriteJSON(map[string]interface{}{"Status": "error", "Message": errAdd.Error()})
		return
	}

	for _, v := range info.ClusterIds {
		err = ConfCreateByClusterId(repo, v.ID)
		if err != nil {
			w.WriteJSON(map[string]interface{}{"Status": "error", "Message": err.Error()})
			return
		}
	}

	var journal model.Journal
	journal.Title = mod.Name
	journal.Operation = "update"
	journal.Resource = "cert"
	journal.Content = "[update cert:" + string(content) + "]"
	journal.User = user.Username
	journal.UpdatedAt = time.Now()
	err = repo.AddJournal(journal)
	if err != nil {
		message = message + fmt.Sprintf("[AddJournal failed:%s]", err)
	}

	if message != "" {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": message})
	} else {
		w.WriteJSON(map[string]interface{}{"Status": "success", "Message": "操作成功"})
	}
}

func timeSub(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.Local)
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, time.Local)
	return int(t1.Sub(t2).Hours() / 24)
}
