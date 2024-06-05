package function

//缓存交互的接口设计
type CacheFunction interface {
	Get() (json string, err error)
	//存储的有效期需要在实现方法的时候自行控制
	Set(json string) error
}
