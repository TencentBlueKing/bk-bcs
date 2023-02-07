package cmd

import (
	"fmt"

	"bscp.io/cmd/data-service/db-migration/migrator"
	"bscp.io/pkg/cc"
	// run the init function to add migrations
	_ "bscp.io/cmd/data-service/db-migration/migrations"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "database migrations tool",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new empty migrations file",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Unable to read flag `name`, err:", err)
			return
		}

		if err := migrator.Create(name); err != nil {
			fmt.Println("Unable to create migration, err:", err)
			return
		}
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "run up migrations",
	Run: func(cmd *cobra.Command, args []string) {

		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`, err:", err)
			return
		}

		if err := cc.LoadSettings(SysOpt); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		db, err := migrator.NewDB()
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		migrator, err := migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		err = migrator.Up(step)
		if err != nil {
			fmt.Println("Unable to run `up` migrations, err:", err)
			return
		}

	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "run down migrations",
	Run: func(cmd *cobra.Command, args []string) {

		step, err := cmd.Flags().GetInt("step")
		if err != nil {
			fmt.Println("Unable to read flag `step`, err:", err)
			return
		}

		if err := cc.LoadSettings(SysOpt); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		db, err := migrator.NewDB()
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		migrator, err := migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		err = migrator.Down(step)
		if err != nil {
			fmt.Println("Unable to run `down` migrations, err:", err)
			return
		}
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "display status of each migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cc.LoadSettings(SysOpt); err != nil {
			fmt.Println("load settings from config files failed, err:", err)
			return
		}

		db, err := migrator.NewDB()
		if err != nil {
			fmt.Println("Unable to new db migrator, err:", err)
			return
		}

		migrator, err := migrator.Init(db)
		if err != nil {
			fmt.Println("Unable to fetch migrator, err:", err)
			return
		}

		if err := migrator.MigrationStatus(); err != nil {
			fmt.Println("Unable to fetch migration status, err:", err)
			return
		}

		return
	},
}

func init() {
	// Add "--name" flag to "create" command
	migrateCreateCmd.Flags().StringP("name", "n", "", "Name for the migration")

	// Add "--step" flag to both "up" and "down" command
	migrateUpCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")
	migrateDownCmd.Flags().IntP("step", "s", 0, "Number of migrations to execute")

	// Add "create", "up" and "down" commands to the "migrate" command
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateCreateCmd, migrateStatusCmd)

	// Add "migrate" command to the root command
	rootCmd.AddCommand(migrateCmd)
}
