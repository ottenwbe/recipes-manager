/*
 * MIT License
 */

package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ottenwbe/recipes-manager/config"
)

var _ = Describe("Config", func() {

	It("is initialized after starting...", func() {
		Expect(config.Config).ToNot(BeNil())
	})

	Context("Viper Configuration", func() {
		It("can read string values from files with arbitrary name and path", func() {
			c := config.NewViperConfig("test-config", []string{"fixtures"})

			s := c.GetString("str")

			Expect(s).To(Equal("success"))
		})

		It("can read integer values from files with arbitrary name and path", func() {
			c := config.NewViperConfig("test-config", []string{"fixtures"})
			i := c.GetInt64("int")
			Expect(i).To(Equal(int64(123)))
		})

		It("can handle string default values", func() {
			const expected = "default"
			const testKey = "default-str"
			c := config.NewViperConfig("test-config", []string{"fixtures"})
			c.SetDefault(testKey, expected)
			s := c.GetString(testKey)
			Expect(s).To(Equal(expected))
		})

		It("can handle int default values", func() {
			const expected = int64(1023)
			const testKey = "default-int"
			c := config.NewViperConfig("test-config", []string{"fixtures"})
			c.SetDefault(testKey, expected)
			i := c.GetInt64(testKey)
			Expect(i).To(Equal(expected))
		})

		It("can handle boo default values", func() {
			const expected = true
			const testKey = "default-bool"
			c := config.NewViperConfig("test-config", []string{"fixtures"})
			c.SetDefault(testKey, expected)
			i := c.GetBool(testKey)
			Expect(i).To(Equal(expected))
		})

		It("can read values from environment variables with correct prefix and separator", func() {
			// Config sets prefix "go_cook" and replaces "." with "_"
			key := "some.env.var"
			envName := "GO_COOK_SOME_ENV_VAR"
			expected := "found_it"

			_ = os.Setenv(envName, expected)
			defer func() { _ = os.Unsetenv(envName) }()

			c := config.NewViperConfig("non-existent", []string{})

			Expect(c.GetString(key)).To(Equal(expected))
		})

		It("should allow manual binding of environment variables", func() {
			key := "manual.bind"
			envName := "GO_COOK_MANUAL_BIND"
			expected := "manual_value"

			_ = os.Setenv(envName, expected)
			defer func() { _ = os.Unsetenv(envName) }()

			c := config.NewViperConfig("non-existent", []string{})
			c.BindEnv(key)

			Expect(c.GetString(key)).To(Equal(expected))
		})

		It("should continue with defaults if config file is not found", func() {
			c := config.NewViperConfig("non-existent-config-file", []string{"."})
			c.SetDefault("foo", "bar")
			Expect(c.GetString("foo")).To(Equal("bar"))
		})
	})
})
