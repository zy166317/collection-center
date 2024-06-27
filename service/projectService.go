package service

import (
	"collection-center/internal/ecode"
	"collection-center/internal/logger"
	"collection-center/internal/rpc"
	"collection-center/library/constant"
	"collection-center/library/request"
	"collection-center/library/utils"
	"collection-center/service/db/dao"
	"math/big"
)

// CheckCollectInfo 校验收款信息是否合法。
// req 包含创建项目的请求信息，其中 CollectInfo 包含收款信息，CollectAddress 包含钱包地址。
// 如果收款信息或钱包地址不合法，函数将返回相应的错误。
// CheckCollectInfo 校验收款信息
func CheckCollectInfo(req *request.CreateProjectReq) error {
	// 检查传入参数是否为空
	if len(req.CollectInfo) <= 0 || len(req.CollectAddress) <= 0 || req.Name == "" || req.Domain == "" || req.NotifyUrl == "" {
		return ecode.IllegalParam
	}

	// 校验钱包地址的格式
	err := utils.CheckWalletAddress(req.CollectAddress)
	if err != nil {
		logger.Error("check wallet address error: %v", err)
		return ecode.CollectAddressFormatError
	}

	// 遍历收款信息，检查每种链上的代币信息
	for chain, tokenInfos := range req.CollectInfo {
		// 检查链对应的收款地址是否存在
		if _, has := req.CollectAddress[chain]; !has {
			return ecode.CollectInfoAddressNotMatch
		}

		// 遍历代币信息，检查每个代币的合法性
		for _, tokenInfo := range tokenInfos {
			// 根据链和代币符号获取代币信息
			// 校验token和rate是否合法
			token, err := dao.GetTokenInfoByChainAndTokenSymbol(chain, tokenInfo.TokenSymbol)
			// 检查获取代币信息是否成功
			if err != nil || token == nil {
				return ecode.TokenSymbolNotExist
			}

			// 稳定币的汇率不能更改
			if tokenInfo.Rate != 100 && tokenInfo.TokenSymbol == "USDT" {
				//稳定币不可设置汇率
				return ecode.UsdtRateNotChange
			}

			// 汇率必须大于0
			if tokenInfo.Rate < 0 {
				return ecode.RateMustBePositive
			}
		}
	}

	// 所有校验通过，返回nil
	return nil
}

// CreateProject 创建一个新的项目，并根据请求中的配置初始化该项目的收款信息。
// req: 创建项目的请求参数，包含项目的基本信息和收款信息。
// merchantUid: 商家用户的唯一标识。
// 返回值: 创建成功的项目对象和错误信息。
func CreateProject(req *request.CreateProjectReq, merchantUid int64) (*dao.Project, error) {
	// 检查请求参数是否合法，确保必要的信息没有缺失。
	if req.Name == "" || req.Domain == "" || req.NotifyUrl == "" {
		return nil, ecode.IllegalParam
	}

	// 生成项目唯一标识。
	projectUid := utils.GenerateUid()

	// 根据请求参数和生成的唯一标识，初始化项目对象。
	project := &dao.Project{
		Name:               req.Name,
		Domain:             req.Domain,
		NotifyUrl:          req.NotifyUrl,
		ProjectUid:         projectUid,
		MerchantUid:        merchantUid,
		ProjectStatus:      dao.ProjectStatusNormal,
		ProjectAuditStatus: dao.ProjectAuditStatusPending,
	}

	// 初始化收集配置数组。
	// 构造db collect records
	collectArr := make([]*dao.Collect, 0)

	// 遍历请求中的收款信息，为每个收款创建一个数据库收款记录
	for k, v := range req.CollectInfo {
		for _, record := range v {
			// 根据链和代币符号获取代币信息。
			// 通过chain和symbol获取tokenInfo
			token, _ := dao.GetTokenInfoByChainAndTokenSymbol(k, record.TokenSymbol)

			// 将收集配置信息和代币信息结合，创建收款记录，并添加到数据库收款数组中。
			collectArr = append(collectArr, &dao.Collect{
				MerchantUid:     merchantUid,
				ProjectUid:      project.ProjectUid,
				CollectUid:      utils.GenerateUid(),
				Chain:           token.Chain,
				TokenSymbol:     token.TokenSymbol,
				ContractAddress: token.ContractAddress,
				Decimals:        token.Decimals,
				LogoUrl:         token.LogoUrl,
				RpcUrl:          token.RpcUrl,
				Rate:            record.Rate,
				CollectAddress:  req.CollectAddress[k],
			})
		}
	}

	// 将项目对象和收款信息数组保存到数据库。
	newProject, _, err := dao.CreateProject(project, collectArr)
	if err != nil {
		// 如果保存过程中发生错误，返回错误信息。
		return nil, ecode.CreateProjectError
	}

	// 返回创建成功的项目对象。
	return newProject, nil
}

// AddTokenInfo 根据请求参数添加新的代币信息到数据库。
// 如果代币存在于以太坊链上，将从链上获取代币的名称、符号和精度。
// req  包含代币信息的请求结构体。
// error - 如果操作失败，返回相应的错误。
func AddTokenInfo(req *request.AddTokenInfoReq) error {
	// 检查请求参数是否为空
	if req.Chain == "" || req.ContractAddress == "" || req.LogoUrl == "" {
		return ecode.IllegalParam
	}

	// 初始化代币的精度、名称和符号变量
	var decimals *big.Int
	var tokenName string
	var tokenSymbol string

	// 如果代币链是以太坊，从链上获取代币信息
	// 检查代币合约地址是否存在
	if req.Chain == constant.EthChain {
		ethRpc, err := rpc.NewEthRpc()
		if err != nil {
			return ecode.CheckTokenAddressError
		}
		decimals, tokenName, tokenSymbol, err = ethRpc.GetTokenInfo(req.ContractAddress)
		if err != nil {
			return ecode.CheckTokenAddressError
		}
	}

	// 创建新的代币信息对象，并填充数据
	// 添加新的代币信息到db
	newToken := &dao.TokenInfo{
		Chain:           req.Chain,
		ContractAddress: req.ContractAddress,
		Decimals:        int(decimals.Int64()),
		LogoUrl:         req.LogoUrl,
		TokenName:       tokenName,
		TokenSymbol:     tokenSymbol,
		RpcUrl:          "", //TODO,后期按配置填写
	}

	// 将新的代币信息插入数据库
	err := dao.CreateTokenInfo(newToken)
	if err != nil {
		return ecode.AddTokenInfoFailed
	}

	// 操作成功，返回nil
	return nil
}

// UpdateProjectInfo 更新项目信息。
// req: 包含需要更新的项目域名和通知URL等信息。
// merchantUid: 商家用户ID。
// 返回错误信息，更新失败或参数不合法。
func UpdateProjectInfo(req *request.UpdateProjectInfo, merchantUid int64) error {
	// 基本参数校验
	//基本参数校验
	if req.Domain == "" || req.NotifyUrl == "" {
		return ecode.IllegalParam
	}
	rows, err := dao.UpdateProjectInfo(merchantUid, req.ProjectUid, req.Domain, req.NotifyUrl)
	// 检查更新是否成功
	if err != nil || rows == 0 {
		return ecode.UpdateProjectInfoFailed
	}
	return nil
}

// UpdateCollectRate 更新汇率。
// req: 包含项目ID、收款ID和新的汇率等信息。
// merchantUid: 商家用户ID。
// 返回错误信息，更新失败或参数不合法。
func UpdateCollectRate(req *request.UpdateCollectRate, merchantUid int64) error {
	// 参数校验
	//参数校验
	if req.ProjectUid == 0 || req.CollectUid <= 0 || merchantUid <= 0 || req.Rate <= 0 {
		return ecode.IllegalParam
	}
	rows, err := dao.UpdateCollectRate(merchantUid, req.ProjectUid, req.CollectUid, req.Rate)
	// 检查更新是否成功
	if err != nil || rows == 0 {
		return ecode.UpdateCollectRateFailed
	}
	return nil
}

// UpdateCollectAddress 更新收款地址。
// req: 包含项目ID、链名称和新的收款地址等信息。
// merchantUid: 商家用户ID。
// 返回错误信息，更新失败或参数不合法。
func UpdateCollectAddress(req *request.UpdateCollectAddress, merchantUid int64) error {
	// 参数校验
	//参数校验
	if req.ProjectUid == 0 || req.Chain == "" || req.Address == "" {
		return ecode.IllegalParam
	}
	rows, err := dao.UpdateCollectsByMerchantUidAndProjectUid(merchantUid, req.ProjectUid, req.Chain, req.Address)
	// 检查更新是否成功
	if err != nil || rows == 0 {
		return ecode.UpdateCollectAddressFailed
	}
	return nil
}

// FreezeProject 根据请求冻结项目。
// req: 冻结项目请求，包含需要冻结的项目UID。
// merchantUid: 商家用户ID，用于标识操作人。
// 返回错误信息，如果操作成功则返回nil。
func FreezeProject(req *request.FreezeProjectReq, merchantUid int64) error {
	// 检查请求中的项目UID是否合法，如果不合法则返回非法参数错误。
	// 参数校验
	if req.ProjectUid == 0 {
		return ecode.IllegalParam
	}

	// 调用DAO层方法冻结项目，传入商家用户ID和项目UID。
	// 冻结项目
	rows, err := dao.FreezeProject(merchantUid, req.ProjectUid)
	// 检查操作是否成功，如果出现错误或者没有影响任何行则返回冻结项目失败错误。
	if err != nil || rows == 0 {
		return ecode.FreezeProjectFailed
	}

	// 如果操作成功，则返回nil。
	return nil
}
