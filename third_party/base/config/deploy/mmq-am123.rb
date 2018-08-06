role :app, "am123" # sandbox
role :db,  "am123", :primary => true # This is where Rails migrations will run
role :web, "am123"                          # Your HTTP server, Apache/etc
set :user, 'qboxserver'
set :repository,  "http://ci.qbox.me/jenkins/job/#{application}-production"
set :deploy_to do "/home/#{user}/websites/mmq" end
set :jenkins_artifact_file, 'dist/qboxmmq'
set :conf_name, 'qboxmmq'


namespace :deploy do
  task :restart, :roles => :app, :max_hosts => 1 do
    run 'supervisorctl restart mmq'
  end
end


